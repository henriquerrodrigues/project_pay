DIARIO DE BORDO
- anotações diárias de progresso do projeto
- Formato padronizado em 24/07/2026 (revisão pós dia 12). Numeração de dia = contador de sessões de trabalho (bate com os "Dia N" do PLANO_MESTRE.md quando o ritmo é 1/dia; dias sem sessão ficam sem entrada, não são renumerados pra "esconder" o buraco).
- Campo `TCC`: sim = candidata forte a virar parágrafo/seção do TCC; talvez = pode virar nota de rodapé ou contexto; não = só operacional.

---

### DIA 01 (LEARN — 2026-05-01)
commit: aed3bd3

**Feito:** pesquisei como funciona o SPI, o DICT, mensagens ISO 20022; aprendi conceitos EndToEndId e protocolos que vêm por meio de XML (vou usar JSON). Modelei as entidades no drawSQL e a relação entre as tabelas.

**Decisão de design:** ISO 20022 usa XML nativamente; optei por JSON no projeto por simplicidade de integração HTTP — não é o padrão real do SPI, é uma simplificação deliberada.

**TCC:** talvez — bom pra justificar, na introdução/metodologia, por que o projeto simplifica XML→JSON sem perder a semântica do domínio (EndToEndId, ISPB, etc.).

**Amanhã:** máquina de estados das transferências.

---

### DIA 02 (BUILD — 2026-05-02)
commit: 9d94b25, 8fd1a7a

**Feito:** criada a estrutura de pastas do repositório; criado `docker-compose` com Postgres 16 e `.env` com as credenciais (adicionado ao `.gitignore`). Essas credenciais são geradas automaticamente pelo Docker ao rodar `docker-compose up -d` — o Postgres inicializa com elas. Criada a conexão com o Postgres.

**TCC:** não — puramente operacional (setup de ambiente).

**Amanhã:** testar a conexão.

---

### DIA 03 (BUILD — 2026-05-03)
commit: bcd4933

**Feito:** testada a conexão com o Postgres usando `pgx` e `conn.Ping()`.

**Erros e como corrigi:** a porta 5432 estava ocupada por uma versão anterior do Postgres instalada localmente → migrei o mapeamento externo do Compose para a porta 5433. Dentro da rede Docker, o host do Postgres é sempre `postgres` e a porta interna continua 5432, independente do mapeamento externo.

**TCC:** não — troubleshooting de ambiente, mas útil registrar como nota de reprodutibilidade (README/"como rodar").

**Amanhã:** migrations.

---

*(sem sessão registrada em 2026-05-04)*

---

### DIA 05 (BUILD — 2026-05-05)
commit: 84965e4

**Feito:** instalado o `golang-migrate` (CLI):
```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```
Criados os arquivos de migration:
```
migrate create -ext sql -dir migrations -seq create_users
migrate create -ext sql -dir migrations -seq create_participants
migrate create -ext sql -dir migrations -seq create_accounts
migrate create -ext sql -dir migrations -seq create_transfers
migrate create -ext sql -dir migrations -seq create_ledger_entries
migrate create -ext sql -dir migrations -seq create_pix_keys
```

**Aprendido — o que é o golang-migrate e por que é útil:**
- Arquivos que controlam a evolução do banco de dados ao longo do tempo; o `golang-migrate` executa os SQLs em ordem e guarda numa tabela interna (`schema_migrations`) qual foi a última migration executada — sabe de onde continuar.
- Sem migration: toda vez que o banco resetar (`docker-compose down -v`) seria preciso recriar as tabelas na mão.
- Com migration: um comando recria tudo (`migrate ... up`); cada mudança vira uma migration numerada, e o histórico fica versionado no Git.
- `.up.sql` aplica a mudança (create table); `.down.sql` desfaz (drop table).
- Reversão parcial: `migrate ... down 2` desfaz as 2 últimas migrations; `migrate ... up` reaplica.
- Comando usado para aplicar:
```
migrate -path migrations -database "postgresql://pay_user:pay_pass@localhost:5433/project_pay?sslmode=disable" up
```
- Onde ver as tabelas: `project_pay → Schemas → public → Table`.

**Decisão de design:** versionamento de schema via `golang-migrate` em vez de `CREATE TABLE` manual/scripts soltos — schema reproduzível e auditável desde o início.

**TCC:** talvez — pode virar uma frase na seção de metodologia/reprodutibilidade ("todo o schema é versionado via golang-migrate, garantindo ambiente idêntico entre desenvolvimento e avaliação").

**Amanhã:** `POST /accounts`, `GET /accounts/{id}`; corrigir `go.mod` (`module main.go` → `module project_pay`); mover a conexão pra `cmd/internal/database/database.go`; `cmd/internal/models/users.go` (struct `Account`); `cmd/internal/repositories/accounts/accounts.go` (queries SQL); `cmd/internal/handlers/accounts.go` (handler HTTP); `cmd/api/main.go` (roteador e servidor).

---

*(sem sessão registrada em 2026-05-06)*

---

### DIA 07 (BUILD — 2026-05-07)
commit: 66651f5

**Feito:** corrigido o nome do módulo em `go.mod`; iniciado `cmd/internal/models/users.go`.

**TCC:** não.

---

### DIA 08 (BUILD — 2026-05-15)
commit: fcb1ca8

**Feito:** revisadas as estruturas e métodos criados para `users`; criadas as estruturas e métodos para `accounts`.

**Erros e como corrigi:** eu só sabia montar consultas com `database/sql`; aprendi a montar com `pgx`. Achava `database/sql` mais organizado (posso migrar de volta depois — decisão em aberto). Em Postgres, o equivalente ao `GETDATE()` do SQL Server é `NOW()`.

**Decisão de design (em aberto):** `pgx` vs `database/sql` — escolhido `pgx` por ser mais idiomático/eficiente pra Postgres especificamente, mas a preferência pessoal por `database/sql` ainda não tem um "porquê" formalizado.

**TCC:** sim — essa é literalmente a ADR-0001 já registrada em `docs/adr/`. Se ainda não tiver o "porquê" (não só a preferência), vale fechar esse raciocínio na própria ADR — é pergunta óbvia de banca ("por que pgx?").

**Ficou para amanhã:** `participants`, `ledger_entries`, `pix_keys`, handlers padrão; corrigir o diagrama com as relações corretas.

---

### DIA 09 (BUILD — 2026-05-16)
commit: 02c6365, 1dcc7e0, 8628d0b, f230339

**Feito:** métodos e structs de `participants`, `pix_keys`, `ledger_entries`, e início de `transfers`; ajuste nas estruturas de `participants`.

**Aprendido:** ISPB = Identificador do Sistema de Pagamentos Brasileiro — código de 8 dígitos que o Banco Central atribui a cada instituição participante do sistema de pagamentos. É o "CPF do banco".

**TCC:** talvez — o ISPB é peça central do domínio real do PIX; vale uma nota no capítulo de fundamentação teórica sobre o SPI.

---

### DIA 10 (BUILD — 2026-05-17)
commit: 4317398

**Feito:** finalizada a camada completa de `transfers`: model, repository com transação atômica, handler e rotas no `main.go`. Implementada a função `generateEndToEndID` no repository — formato `E + ISPB(8 dígitos) + YYYYMMDD + HHMM + sufixo aleatório(11 chars)`, exemplo: `E60074319020260517142300A3F9C12D8`. Implementado `ErrInsufficientFunds` como erro exportado, permitindo o handler distinguir saldo insuficiente (422) de erro interno (500).

**Fluxo da transação atômica:**
1. `SELECT balance + ispb` do remetente com `FOR UPDATE` (trava a linha)
2. Verifica `saldo >= amount`
3. Gera o `end_to_end_id`
4. Debita remetente (`UPDATE ... RETURNING balance`)
5. Verifica se a conta recebedora existe
6. Credita recebedor (`UPDATE ... RETURNING balance`)
7. `INSERT` em `transfers`
8. `INSERT` `ledger_entry` DEBIT (remetente)
9. `INSERT` `ledger_entry` CREDIT (recebedor)
10. `COMMIT`

Registradas as rotas `POST /transfers` e `GET /transfers/{id}` no `main.go`.

**Lições aprendidas:**
- `transfers` é o único recurso que precisa de transação SQL, pois envolve múltiplas operações que precisam ser atômicas.
- `FOR UPDATE` trava a linha do remetente durante a transação para evitar débitos simultâneos (race condition).
- Erros de negócio devem ser erros exportados (`var ErrInsufficientFunds = errors.New(...)`) para o handler conseguir tratá-los com `errors.Is()` e retornar o status HTTP correto.
- `tx.Exec` para operações sem retorno (ledger entries); `tx.QueryRow` quando precisa do `RETURNING`.
- `defer tx.Rollback()` é seguro chamar mesmo depois do `Commit` — `pgx` ignora se a transação já foi encerrada.

**Erros e como corrigi:**
- `CompletedAt` era `time.Time`, mas deveria ser `*time.Time`, pois o campo é nullable no banco — quando `NULL`, o scan falhava.
- Removidos `EndToEndID` e `Status` de `CreateTransferRequest`: não devem vir do cliente, são gerados/controlados pelo sistema.
- `sqlCreditAccount` era chamado com 1 argumento, mas o SQL precisava de 2 (`$1=amount`, `$2=id`) — passava apenas o `id`.
- `GetByID` retornava `transfer, nil` em vez de `transfer, err` após o scan.
- Faltava verificação `if err != nil` após o scan do recebedor — um erro de banco passaria silencioso.

**TCC:** sim — este é o dia mais denso do diário em conteúdo de TCC até aqui. `FOR UPDATE` prevenindo race condition é exatamente o tipo de decisão que vira seção de "consistência sob concorrência" (e vai ser reexaminada/comparada na Fase 2 com locking otimista). A lista de "erros e como corrigi" também é material direto pra metodologia (mostra processo de depuração real).

---

### RESUMO — o que foi implementado até o Dia 10

No primeiro dia aprendi regras de negócio como SPI, DICT e mensagens ISO 20022; modelei as entidades no drawSQL e a relação entre as tabelas; fiz a máquina de estados do domínio; criei as pastas e arquivos do repositório `project_pay`; criei as tabelas com `golang-migrate` (recriadas automaticamente via `docker-compose`); criei a conexão com o Postgres usando `pgx`; e implementei todos os endpoints iniciais do projeto, incluindo as regras de validação para transferências atômicas.

---

### DIA 11 (REFACTOR/DOC — 2026-05-21)
commit: 88fada4

**Feito:** revisão do que foi criado até aqui; testada com sucesso a primeira transferência via `POST /transfers`.

**Lições devops:**

pgAdmin — acessar:
```
URL:   http://localhost:5050
Email: admin@admin.com
Senha: admin
```

Configuração do servidor no pgAdmin (fazer uma vez após `down -v`):
```
Aba General    → Name: project_pay
Aba Connection → Host: host.docker.internal
                 Port: 5432
                 Database: project_pay
                 Username: pay_user
                 Password: pay_pass
```

Banco — comandos úteis via Docker:
```bash
# acessar o banco direto pelo terminal
docker exec -it project_pay_db psql -U pay_user -d project_pay

# dentro do psql
\dt                  # listar tabelas
\d nome_da_tabela    # ver estrutura da tabela
\q                   # sair
```

Ordem para iniciar do zero:
```
1. PowerShell como admin → parar serviços PostgreSQL locais
2. docker-compose up -d
3. migrate ... up
4. go run ./cmd/api/.
5. Postman → testar endpoints
```

**TCC:** não — runbook operacional. Boa base pro README de "como rodar", mas não é conteúdo de tese.

---

### DIA 12 (BUILD — 2026-05-22)
commit: d657d1e *(nota: commit registrado em 2026-05-24 — checar se a sessão de fato ocorreu em 22/05 com push tardio, ou se a data desta entrada está errada)*

**Feito:** criada autenticação JWT (`Auth`), por meio de tokens Bearer.

**TCC:** talvez — depende de quanto a discussão de segurança/autenticação for aprofundada no trabalho; se o foco do TCC for consistência/idempotência (tema 1 sugerido no plano mestre), isso fica como contexto, não como capítulo central.

---

### DIA 13 (DOC/PLAN — 2026-07-13)
commit: 2268ad2, a9f76a2, a727776, 5311422, 4b13103, 2afa1f9, 4a0f0ec, 2d7afb8, 5dabd33, 2a3aad9, b02e39b

**Feito:** retomada do projeto após um hiato longo (~7 semanas desde o Dia 12, em 24/05). Dia dedicado a planejamento e automação, não a código de produto:
- Criados os arquivos de processo `DIARIO_DE_BORDO.md`, `PLANO_MESTRE.md` (plano de 12 meses) e `RITUAL_DIARIO.md`.
- Escritas as specs e planos detalhados das Semanas 3, 4 e 5 (`docs/superpowers/specs/` e `docs/superpowers/plans/`) — autenticação/erros/camadas, frontend mínimo, Swagger/`.http`/README/ADR.
- Configurados os workflows do GitHub Actions "Claude PR Assistant" e "Claude Code Review".
- Ajuste de licença no README.

**Decisão de design:** formalizar o plano mestre de 12 meses e o diário de bordo como artefatos versionados do próprio repositório, não como anotações soltas — torna o processo de aprendizagem parte auditável do projeto (relevante pro TCC).

**TCC:** talvez — o próprio `PLANO_MESTRE.md` e a existência de um diário de bordo estruturado são evidência de metodologia; vale citar na seção de metodologia do TCC como "processo de desenvolvimento orientado a plano e diário".

**Observação:** existe um hiato de ~50 dias sem commit em `main`/`master` entre o Dia 12 (24/05) e este dia (13/07). Vale registrar o motivo real (ex.: período de provas, férias, outra prioridade) — é dado legítimo pro capítulo de metodologia ("gestão do tempo em projeto de longo prazo").

---

### DIA 14 (BUILD — 2026-07-21)
commit: bb2a759

**Feito:** iniciada a fundação do contrato de erro padronizado — pacote `apperrors` (tipo `Error`, construtores, `WriteHTTP`, testes), pacote `traceid`, e o helper `middleware.UserIDFromContext`.

**Decisão de design:** erro de aplicação tipado que atravessa handler/service/repository sem vazar detalhe interno; trace ID isolado em pacote próprio pra evitar ciclo de import entre `middleware` e `apperrors`.

**TCC:** sim — é a base do capítulo de "tratamento de erros e observabilidade" (contrato de erro consistente, correlação por trace ID).

---

### DIA 15 (BUILD — 2026-07-22)
commit: eed3701, 8d4f0d6

**Feito:** camada de services iniciada.
- `services/users` completo: interface `UserRepository`, `Create` (validação de campos obrigatórios), `GetByID` (ownership — 404 vs 403), testes com fake repository em memória; handler e `main.go` religados pro service.
- `services/accounts` iniciado: interface `AccountRepository`, `ListByCurrentUser` implementado; `Create`/`GetByID` ainda pendentes. Repository de `accounts` ajustado (`Create` recebe `userID`; novo `ListByUserID`); `CreateAccountRequest` não aceita mais `UserID` no body (dono vem do token).

**Decisão de design:** interface do repository declarada dentro do próprio pacote do service (não em `repositories/`), permitindo fakes em memória nos testes sem depender de banco real — inversão de dependência aplicada de forma manual (sem framework de DI).

**TCC:** sim — inversão de dependência manual + ownership (404 vs 403) é conteúdo direto pra "arquitetura em camadas e testabilidade" no TCC; a remoção de `UserID` do body de `CreateAccountRequest` também é um bom exemplo de correção de falha de autorização (IDOR) pra citar em segurança.

**Amanhã:** fechar `Create`/`GetByID` de `services/accounts`, depois replicar o padrão em `participants`, `pix_keys`, `ledger_entries`, `transfers` e `auth`.
