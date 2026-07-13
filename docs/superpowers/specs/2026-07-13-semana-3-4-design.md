# Spec — Semanas 3 e 4 do Plano Mestre

> Data: 2026-07-13
> Escopo: Semana 3 (autenticação/erros/logs/camadas) e Semana 4 (frontend mínimo) da Fase 1.
> Referência: `.github/workflows/plans/PLANO_MESTRE.md`

## Contexto e estado atual

- JWT já implementado: `POST /auth/login`, middleware `Auth`, bcrypt, claims com `user_id` (`cmd/internal/repositories/authentication/auth.go`, `cmd/internal/middleware/auth.go`). **Mantém como está** — fora de escopo revisar.
- Faltam da Semana 3: pacote de erros (stub em `cmd/internal/models/errors.go`), logs estruturados (main usa `log`), camada de services (`cmd/internal/services/services.go` vazio — handlers chamam repositories direto), validação de ownership.
- Semana 4 não começada: nenhum frontend no repo.

## Decisões tomadas

| Decisão | Escolha |
|---|---|
| Escopo de auth | Manter JWT atual; adicionar apenas ownership |
| Camadas | Service em **todas** as entidades, mesmo CRUD simples |
| Auth no frontend | Cookie httpOnly via route handlers do Next (proxy) |
| Ordem de execução | Backend inteiro (Semana 3) primeiro, depois frontend (Semana 4) |
| Ownership negado | Retorna **403 Forbidden** (não 404) |

---

## Parte 1 — Backend (Semana 3)

### 1.1 Pacote `apperrors` (novo: `cmd/internal/apperrors/`)

- Tipo `Error` com campos:
  - `Code string` — código estável de máquina (ex.: `INSUFFICIENT_FUNDS`, `NOT_FOUND`, `FORBIDDEN`, `INVALID_INPUT`, `UNAUTHORIZED`, `INTERNAL`).
  - `HTTPStatus int` — status HTTP correspondente.
  - `Message string` — mensagem segura para o cliente (pt-BR, sem detalhe interno).
  - `err error` — causa interna embrulhada (`Unwrap()`); **nunca** serializada no response.
- Construtores: `NotFound(msg)`, `InvalidInput(msg)`, `Unauthorized(msg)`, `Forbidden(msg)`, `InsufficientFunds()`, `Internal(err)`.
- `WriteHTTP(w http.ResponseWriter, r *http.Request, err error)`:
  - Se `err` é `*apperrors.Error`: responde JSON `{"code": "...", "message": "...", "trace_id": "..."}` com o `HTTPStatus`.
  - Qualquer outro erro: responde 500 com código `INTERNAL` e mensagem genérica; detalhe vai só pro log (com trace_id).
- Substitui **todos** os `http.Error(...)` dos handlers.
- `models/errors.go` (stub `errorResponse`) é removido.

### 1.2 Logs estruturados com `slog`

- `main.go`: `slog.New(slog.NewJSONHandler(os.Stdout, ...))` + `slog.SetDefault`.
- Middleware `RequestLogger` (`cmd/internal/middleware/`):
  - Gera trace ID (UUID v4) por request; injeta no `context` (chave própria tipada) e no header de response `X-Trace-Id`.
  - Ao final do request loga: método, rota, status, duração, trace_id.
  - Aplica em **todas** as rotas (envolve o mux inteiro).
- `apperrors.WriteHTTP` loga o erro interno com trace_id via `slog` — cliente correlaciona pelo `trace_id` do JSON/header.

### 1.3 Camada de services (todas as entidades)

- Estrutura: `cmd/internal/services/<entidade>/` — um pacote por entidade, espelhando `repositories/`.
- Responsabilidades:
  - **Handler:** decode do body, parse de path params, chamar service, escrever response via `apperrors.WriteHTTP` no erro.
  - **Service:** validação de negócio, ownership, orquestração; retorna erros `apperrors`.
  - **Repository:** SQL puro; sem regra de negócio.
- Validações migram dos handlers para os services (ex.: `amount <= 0`, origem == destino, hoje em `handlers/transfers.go`).
- Cada service depende de uma **interface** pequena do repository (definida no pacote do service) — permite fake nos testes.
- `main.go` monta: repo → service → handler.
- Arquivos placeholder vazios (`services/services.go`, `repositories/repositories.go`, `handlers/handlers.go`) removidos se não ganharem conteúdo.

### 1.4 Ownership

- Services leem `user_id` do context (chave exportada pelo middleware `Auth`).
- Regras:
  - `GET /accounts/{id}` e `GET /accounts/{id}/statement`: conta deve pertencer ao user → senão 403.
  - `GET /transfers/{id}`: user deve ser dono da conta origem **ou** destino → senão 403.
  - `POST /transfers`: conta origem deve pertencer ao user → senão 403.
  - `POST /accounts`: conta criada vinculada ao `user_id` do token; body **não** aceita `user_id`.
  - `GET /users/{id}`: só o próprio user → senão 403.
- Recurso inexistente continua 404; existente mas de outro user, 403.

### 1.5 Testes

- Unitários de services com repository fake: ownership (caso permitido e caso 403), validações de negócio, mapeamento de erros.
- Teste do middleware `RequestLogger` (trace ID presente no context e no header).
- Teste do `apperrors.WriteHTTP` (erro conhecido → JSON correto; erro desconhecido → 500 genérico sem vazar detalhe).

---

## Parte 2 — Frontend (Semana 4)

### 2.1 Setup

- Diretório `web/` na raiz (monorepo).
- Next.js 15 (App Router) + TypeScript + Tailwind v4 + shadcn/ui.
- API Go acessada server-side via `API_URL` em `web/.env.local`. `docker-compose` não muda; front roda `npm run dev`.

### 2.2 Integração com API — cookie httpOnly

- Browser **nunca** vê o JWT. Route handlers do Next fazem proxy:
  - `POST /api/login`: repassa para `POST /auth/login` do Go; sucesso grava JWT em cookie httpOnly (`secure` em produção, `sameSite=lax`, expiração alinhada às 24h do token).
  - `POST /api/logout`: apaga o cookie.
- Helper server-side `apiFetch(path, init?)`: lê o cookie e injeta `Authorization: Bearer <jwt>` nas chamadas à API Go. Server Components buscam dados com ele.
- Mutações (transferência, cadastro) via Server Actions ou route handlers — mesma mecânica de cookie.
- `middleware.ts` do Next: rotas protegidas sem cookie → redirect para `/login`.

### 2.3 Telas

- `/login` — email + senha; erro de credencial exibido no form.
- `/signup` — cria user via `POST /users` (rota pública); sucesso → redirect para `/login`.
- `/dashboard` — lista contas do user com saldo; Dialog (shadcn) de transferência: conta origem (select das contas do user), conta destino, valor; sucesso → refresh dos saldos.
- `/accounts/[id]/statement` — extrato em tabela shadcn, paginação "carregar mais".
- Erros da API: exibir campo `message` do JSON de erro (contrato da Parte 1).
- Server Components por padrão; client components só onde há interação (forms, dialog).

### 2.4 Critério de aceite (marco da Semana 4)

Fluxo completo no navegador: cadastrar → logar → ver contas e saldos → transferir entre contas → ver extrato refletindo a transferência. Estético não importa; funcional importa.

---

## Fora de escopo

- Refresh token / revogação de JWT.
- Revisão da implementação JWT existente.
- Testes E2E de frontend.
- Responsividade caprichada / dark mode.
- Deploy do frontend.
- Swagger (Semana 5), testes de carga (Semana 6).

## Riscos

- Refatoração em camadas toca todos os handlers — fazer entidade por entidade, com `go build` + testes verdes a cada passo.
- Contrato de erro muda o formato de response de erro — frontend (Parte 2) já nasce consumindo o formato novo; nenhum consumidor externo existe ainda.
