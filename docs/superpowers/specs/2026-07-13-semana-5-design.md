# Spec — Semana 5 do Plano Mestre (Documentação e Swagger)

> Data: 2026-07-13
> Escopo: Semana 5 da Fase 1 — Swagger via swaggo, arquivos .http, README, primeira ADR.
> Referência: `.github/workflows/plans/PLANO_MESTRE.md`

## Contexto e premissa

- **Pré-requisito:** Semanas 3 e 4 implementadas antes desta (spec em `2026-07-13-semana-3-4-design.md`). O Swagger documenta a API **pós-refatoração**: camadas handler/service/repository, contrato de erro `apperrors` (`{"code", "message", "trace_id"}`), JWT com ownership, frontend Next em `web/`.
- Endpoints a documentar (11):
  - Públicos: `POST /auth/login`, `POST /users`.
  - Protegidos: `GET /users/{id}`, `POST /accounts`, `GET /accounts/{id}`, `GET /accounts/{id}/statement`, `POST /transfers`, `GET /transfers/{id}`, `POST /participants`, `GET /participants/{id}`, `POST /pix-keys`, `GET /pix-keys/{key}`.

## Decisões tomadas

| Decisão | Escolha |
|---|---|
| Pré-requisito | Assume Semanas 3-4 prontas; documenta contrato apperrors |
| Coleção de requests | Arquivos `.http` versionados (não Postman) |
| Tema da primeira ADR | pgx em vez de database/sql |
| Diagrama do README | Mermaid inline (renderiza no GitHub) |
| Geração Swagger | `swag init` com pacote gerado **commitado** no repo |

---

## 1. Swagger com swaggo (dias 29-30)

### Anotações gerais (`cmd/api/main.go`)

- `@title Project Pay API`, `@version`, `@description` (motor de pagamentos instantâneos inspirado no PIX).
- `@host localhost:8080` (porta de `API_PORT` no `.env.example`), `@BasePath /`.
- `@securityDefinitions.apikey BearerAuth` — `in: header`, `name: Authorization`. Descrição instrui prefixo `Bearer `.

### Anotações por handler

- Cada endpoint recebe: `@Summary`, `@Description` (quando o summary não basta), `@Tags` (uma tag por entidade: auth, users, accounts, transfers, participants, pix-keys), `@Accept json` / `@Produce json`, `@Param` (path/body), `@Success` com o model de response, `@Failure` para cada erro possível (400, 401, 403, 404, 422, 500 conforme o caso) apontando pro model de erro, `@Router`.
- Rotas protegidas levam `@Security BearerAuth`.
- Model de erro documentado: struct do pacote `apperrors` com `code`, `message`, `trace_id` (exatamente o JSON que `WriteHTTP` serializa). Se a struct de response de erro não for exportada, exportar (ex.: `apperrors.ErrorResponse`) só pro swaggo referenciar.
- Models de request/response existentes em `cmd/internal/models/` ganham `example` tags onde ajudar a UI (ex.: `example:"100.50"` em amount).

### Geração e serving

- `swag init` gera pacote Go em `docs/swagger/` (`--output docs/swagger`, `--parseInternal` se necessário). **Commitado** — quem clona builda sem rodar swag.
- `main.go`: blank import `_ "project_pay/docs/swagger"` + rota `mux.Handle("GET /swagger/", httpSwagger.WrapHandler)` (pacote `github.com/swaggo/http-swagger`).
- Rota `/swagger/` é pública (sem middleware Auth).
- Dependências novas: `github.com/swaggo/swag` (CLI), `github.com/swaggo/http-swagger`.

## 2. Arquivos .http (dia 31)

- Pasta `requests/` na raiz, um arquivo por entidade: `auth.http`, `users.http`, `accounts.http`, `transfers.http`, `participants.http`, `pix_keys.http`.
- Convenções:
  - `@baseUrl = http://localhost:8080` no topo de cada arquivo.
  - Login em `auth.http` como **named request** (`# @name login`); demais arquivos reusam token via `{{login.response.body.token}}` (suportado por REST Client no VS Code e GoLand).
  - Cada arquivo cobre o caminho feliz de todos os endpoints da entidade **+ pelo menos 1 caso de erro** (ex.: transferência com saldo insuficiente) — demonstra o contrato de erro na prática.
- `requests/README.md` curto: como usar (extensão REST Client / GoLand), ordem sugerida (criar user → login → criar contas → transferir).

## 3. README do projeto (dia 32)

Seções, nesta ordem:

1. **Visão** — o que é o Project Pay, objetivo do projeto (motor de pagamentos com foco em concorrência/idempotência), fase atual.
2. **Stack** — Go, Postgres 16 + pgx, Next.js, Docker Compose.
3. **Como rodar** — passo a passo: clonar, copiar `.env.example` pra `.env`, `docker compose up`, migrations, subir API, subir front (`web/`), rodar testes. Meta do marco: outra pessoa roda sem perguntar nada.
4. **Arquitetura** — diagrama **Mermaid** inline: browser → Next (proxy cookie httpOnly) → API Go → camadas handler → service → repository → Postgres. Referencia `docs/diagrama.jpg` (modelo de dados) e `docs/dominio.md`.
5. **Documentação da API** — link pro `/swagger/` local e pra pasta `requests/`.
6. **Estrutura de pastas** — árvore comentada de `cmd/`, `web/`, `docs/`, `requests/`.
7. **Decisões arquiteturais** — link pra `docs/adr/`.
8. **Próximos passos** — Semana 6 (testes de carga) e visão das fases.
- Linha de licença existente preservada.

## 4. Primeira ADR (dias 33-34)

- `docs/adr/README.md`: o que é ADR, índice das ADRs, template.
- Template (seções): **Status** (Aceita/Substituída/Deprecada) / **Contexto** / **Decisão** / **Consequências** (positivas e negativas).
- `docs/adr/0001-pgx-em-vez-de-database-sql.md`:
  - Contexto: acesso a Postgres em Go; alternativas `database/sql` + driver (`lib/pq`, `pgx/stdlib`) vs `pgx` nativo.
  - Decisão: `pgx` nativo (pool `pgxpool`, tipos Postgres nativos, performance, prepared statements automáticos).
  - Consequências: acoplamento a Postgres (aceitável — Postgres é decisão de projeto), API diferente de `database/sql`, melhor suporte a tipos (numeric, timestamptz).
- Escrita retroativa: registra decisão já tomada na Semana 1 — dizer isso no Status/Contexto.

## 5. Verificação (dia 35 — revisão)

- `swag init` roda sem warning relevante; `go build ./...` e testes verdes.
- `/swagger/` abre no browser; botão **Authorize** com token de login funciona; executar `POST /transfers` pela UI com sucesso.
- Cada arquivo `.http` executável de ponta a ponta contra a API local.
- README testado "de fora": seguir os passos numa cabeça limpa (ou máquina limpa) e anotar o que faltou.

## Fora de escopo

- Check de drift do Swagger em CI (Semana 8).
- Versionamento de API (`/v1`).
- Documentação do frontend além do "como rodar".
- Postman collection.
- Testes de carga (Semana 6).

## Riscos

- **Drift silencioso:** anotações swaggo desatualizam quando handlers mudam. Mitigação definitiva na Semana 8 (CI valida `swag init` sem diff); até lá, disciplina manual — atualizar anotação no mesmo commit que muda o handler.
- **Conflito de pasta:** `docs/swagger/` é código gerado dentro de `docs/` (que hoje só tem markdown). Aceito — mantém convenção padrão do swaggo e um único lugar de documentação.
