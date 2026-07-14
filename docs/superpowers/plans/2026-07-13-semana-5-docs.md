# Semana 5 — Documentação e Swagger — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** API documentada de ponta a ponta: Swagger UI servido em `/swagger/` (swaggo, pacote gerado commitado), coleção de arquivos `.http` versionada em `requests/`, README completo com diagrama Mermaid e primeira ADR (`pgx` em vez de `database/sql`).

**Architecture:** Anotações swaggo nos handlers existentes (camadas handler → service → repository da Semana 3). `swag init` gera o pacote em `docs/swagger/` (commitado — quem clona builda sem rodar swag). O contrato de erro `{"code","message","trace_id"}` vira o model exportado `apperrors.ErrorResponse`, referenciado em todo `@Failure`. UUIDs viram `string` no spec via arquivo de overrides `.swaggo`.

**Tech Stack:** Go 1.25, `github.com/swaggo/swag` (CLI), `github.com/swaggo/http-swagger`, mux stdlib (`net/http`), REST Client (VS Code) / HTTP Client (GoLand) pros `.http`.

**Referência:** spec `docs/superpowers/specs/2026-07-13-semana-5-design.md`.

**Pré-requisito:** planos `2026-07-13-semana-3-backend.md` e `2026-07-13-semana-4-frontend.md` executados. Este plano assume o estado pós-Semana 3/4: pacote `apperrors` com `WriteHTTP`, services por entidade, `middleware.Auth`/`RequestLogger`, `GET /accounts`, `POST /users` com campo `password`, `POST /accounts` sem `user_id` no body, frontend em `web/`, seed `scripts/dev_seed.sql`. **Se algum desses não existir, pare e execute os planos anteriores primeiro.**

**Fora de escopo:** check de drift do Swagger em CI (Semana 8), versionamento `/v1`, Postman, testes de carga.

**Adições ao spec (lacunas detectadas):**
1. `GET /accounts` — o spec lista os endpoints originais, mas a Semana 3 adicionou a listagem de contas. Sem anotar, o Swagger já nasce com drift. Incluído — total: **13 operações** documentadas.
2. Token cross-file nos `.http` — request variables do REST Client (`{{login.response.body.$.token}}`) **só funcionam dentro do mesmo arquivo**. Cada `.http` ganha seu próprio bloco de login (pequena duplicação; garante o critério do spec "cada arquivo executável de ponta a ponta").
3. `JWT_SECRET` fora do `.env.example` — o código lê `os.Getenv("JWT_SECRET")` (`cmd/internal/repositories/authentication/auth.go:27`) mas a variável não está no `.env.example`. Quebra a meta "outra pessoa roda sem perguntar nada". Corrigido na Task 11, junto do `DB_PORT` (compose expõe 5433, exemplo diz 5432).
4. `models.LoginResponse` — o login responde `map[string]string{"token": ...}`; swaggo precisa de struct nomeada pro `@Success`. Task 3 (JSON de saída não muda).

---

## Estrutura de arquivos (visão geral)

| Arquivo | Responsabilidade |
|---|---|
| `.swaggo` (novo) | Override: `uuid.UUID` → `string` no spec |
| `docs/swagger/` (gerado, commitado) | `docs.go`, `swagger.json`, `swagger.yaml` |
| `cmd/api/main.go` (modificar) | Anotações gerais `@title`/`@securityDefinitions` + rota `/swagger/` + blank import |
| `cmd/internal/apperrors/apperrors.go` (modificar) | `ErrorResponse` exportada (era `response`) |
| `cmd/internal/models/auth.go` (modificar) | `LoginResponse` + example tags |
| `cmd/internal/models/*.go` (modificar) | Example/enums tags |
| `cmd/internal/handlers/*.go` (modificar) | Anotações swaggo por endpoint |
| `requests/*.http` + `requests/README.md` (novos) | Coleção de requests versionada |
| `README.md` (reescrever) | Visão, stack, como rodar, arquitetura, links |
| `.env.example` (modificar) | `JWT_SECRET`, `DB_PORT=5433`, credenciais dev |
| `docs/adr/README.md` + `docs/adr/0001-pgx-em-vez-de-database-sql.md` (novos) | ADRs |

Comandos rodam da raiz do repo (`c:\Go\repos\project_pay`).

**Comando de geração (usado em várias tasks — sempre o mesmo):**

```bash
swag init -g cmd/api/main.go -o docs/swagger --parseInternal --parseDependency
```

---

### Task 1: Dependências swaggo

**Files:** `go.mod`, `go.sum` (via `go get`)

- [ ] **Step 1: Instalar o CLI `swag`**

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag --version
```

Expected: imprime a versão (ex.: `swag version v1.16.x`). Se `swag` não for encontrado, o `GOPATH/bin` não está no PATH — no PowerShell: `$env:PATH += ";$(go env GOPATH)\bin"`.

- [ ] **Step 2: Adicionar o http-swagger ao módulo**

```bash
go get github.com/swaggo/http-swagger
go build ./...
```

Expected: `go.mod` ganha `github.com/swaggo/http-swagger` (e transitivas `swaggo/files`, `swaggo/swag`); build limpo.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: adiciona dependencias swaggo (http-swagger)"
```

---

### Task 2: Exportar `apperrors.ErrorResponse`

O swaggo precisa referenciar a struct do contrato de erro em `@Failure`; hoje ela é a `response` não exportada dentro de `WriteHTTP`.

**Files:**
- Modify: `cmd/internal/apperrors/apperrors.go`

- [ ] **Step 1: Renomear `response` → `ErrorResponse` com example tags**

Em `cmd/internal/apperrors/apperrors.go`, trocar:

```go
type response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}
```

por:

```go
// ErrorResponse é o JSON do contrato de erro da API — exatamente o que
// WriteHTTP serializa. Exportada para o swaggo referenciar nos @Failure.
type ErrorResponse struct {
	Code    string `json:"code" example:"NOT_FOUND"`
	Message string `json:"message" example:"conta não encontrada"`
	TraceID string `json:"trace_id" example:"6fdd7bfc-4c5a-4f24-8a77-1c1b1b9a4f10"`
}
```

E dentro de `WriteHTTP`, trocar:

```go
	json.NewEncoder(w).Encode(response{Code: appErr.Code, Message: appErr.Message, TraceID: tid})
```

por:

```go
	json.NewEncoder(w).Encode(ErrorResponse{Code: appErr.Code, Message: appErr.Message, TraceID: tid})
```

- [ ] **Step 2: Rodar os testes do pacote (validam o contrato JSON — nada deve mudar)**

Run: `go test ./cmd/internal/apperrors/ -v`
Expected: PASS (todos — o JSON serializado é idêntico).

- [ ] **Step 3: Commit**

```bash
git add cmd/internal/apperrors/apperrors.go
git commit -m "refactor: exporta apperrors.ErrorResponse pro swagger referenciar"
```

---

### Task 3: `models.LoginResponse`

**Files:**
- Modify: `cmd/internal/models/auth.go`
- Modify: `cmd/internal/handlers/auth.go`

- [ ] **Step 1: Adicionar a struct de response e example tags no request**

Substituir `cmd/internal/models/auth.go` por:

```go
package models

// AuthJWT é o body do POST /auth/login.
type AuthJWT struct {
	Email    string `json:"email" example:"fulano@email.com"`
	Password string `json:"password" example:"senha123"`
}

// LoginResponse é o body de sucesso do POST /auth/login.
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
```

- [ ] **Step 2: Handler usa a struct (JSON de saída idêntico)**

Em `cmd/internal/handlers/auth.go`, no final de `Login`, trocar:

```go
	json.NewEncoder(w).Encode(map[string]string{"token": token})
```

por:

```go
	json.NewEncoder(w).Encode(models.LoginResponse{Token: token})
```

(O import `"project_pay/cmd/internal/models"` já existe no arquivo — `Login` decodifica `models.AuthJWT`.)

- [ ] **Step 3: Build + testes**

Run: `go build ./... ; go test ./...`
Expected: tudo verde.

- [ ] **Step 4: Commit**

```bash
git add cmd/internal/models/auth.go cmd/internal/handlers/auth.go
git commit -m "refactor: LoginResponse tipada pro swagger (JSON inalterado)"
```

---

### Task 4: Anotações gerais + `.swaggo` + rota `/swagger/` + primeira geração

**Files:**
- Create: `.swaggo`
- Create: `docs/swagger/` (gerado pelo `swag init`)
- Modify: `cmd/api/main.go`

- [ ] **Step 1: Criar o arquivo de overrides `.swaggo` na raiz**

`uuid.UUID` é um `[16]byte` — sem override, o spec renderiza os IDs como array de inteiros. Conteúdo do arquivo `.swaggo`:

```
replace github.com/google/uuid.UUID string
```

- [ ] **Step 2: Anotações gerais no `main.go`**

Em `cmd/api/main.go`, imediatamente acima de `func main() {`:

```go
// @title           Project Pay API
// @version         1.0
// @description     Motor de pagamentos instantâneos inspirado no PIX. Valores monetários em centavos (int64). Erros seguem o contrato {"code","message","trace_id"}.

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Informe "Bearer " seguido do token JWT obtido no POST /auth/login. Ex.: Bearer eyJhbGci...
func main() {
```

- [ ] **Step 3: Gerar o pacote Swagger**

```bash
swag init -g cmd/api/main.go -o docs/swagger --parseInternal --parseDependency
```

Expected: cria `docs/swagger/docs.go`, `docs/swagger/swagger.json`, `docs/swagger/swagger.yaml` sem erro. Ainda sem paths (nenhum `@Router` anotado) — normal nesta task.

Se o swag reclamar de `uuid.UUID` mesmo com o `.swaggo` (versões antigas ignoram o arquivo), o fallback é adicionar `swaggertype:"string" format:"uuid"` nos campos `uuid.UUID` dos models — mas atualize o swag primeiro (`go install github.com/swaggo/swag/cmd/swag@latest`).

- [ ] **Step 4: Rota `/swagger/` + blank import no `main.go`**

Adicionar aos imports de `cmd/api/main.go`:

```go
	httpSwagger "github.com/swaggo/http-swagger"

	_ "project_pay/docs/swagger"
```

E junto das rotas públicas (perto de `POST /auth/login` — **sem** `middleware.Auth`):

```go
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
```

- [ ] **Step 5: Build + smoke**

Run: `go build ./... ; go vet ./...`
Expected: limpo.

Smoke (precisa do Postgres de pé): `docker compose up -d postgres`, depois `go run ./cmd/api` e abrir `http://localhost:8080/swagger/index.html` — a UI carrega com título "Project Pay API" e zero endpoints.

- [ ] **Step 6: Commit (inclui o pacote gerado)**

```bash
git add .swaggo cmd/api/main.go docs/swagger/
git commit -m "feat: swagger ui em /swagger/ com anotacoes gerais da api"
```

---

### Task 5: Anotar handlers de auth e users

**Files:**
- Modify: `cmd/internal/handlers/auth.go`
- Modify: `cmd/internal/handlers/users.go`
- Regenerate: `docs/swagger/`

- [ ] **Step 1: Anotar `Login`**

Em `cmd/internal/handlers/auth.go`, imediatamente acima de `func (h *AuthHandler) Login(`:

```go
// Login godoc
// @Summary      Autentica um usuário e retorna o token JWT
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      models.AuthJWT  true  "Credenciais"
// @Success      200   {object}  models.LoginResponse
// @Failure      400,401,500  {object}  apperrors.ErrorResponse
// @Router       /auth/login [post]
```

- [ ] **Step 2: Anotar `Create` e `GetByID` de users**

Em `cmd/internal/handlers/users.go`, acima de `func (h *UserHandler) Create(`:

```go
// Create godoc
// @Summary      Cria um usuário
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreateUserRequest  true  "Dados do usuário"
// @Success      201   {object}  models.CreateUserResponse
// @Failure      400,500  {object}  apperrors.ErrorResponse
// @Router       /users [post]
```

Acima de `func (h *UserHandler) GetByID(`:

```go
// GetByID godoc
// @Summary      Busca um usuário por ID (apenas o próprio)
// @Description  404 se não existe; 403 se o ID é de outro usuário.
// @Tags         users
// @Produce      json
// @Param        id  path  string  true  "ID do usuário (UUID)"
// @Success      200  {object}  models.User
// @Failure      400,401,403,404,500  {object}  apperrors.ErrorResponse
// @Security     BearerAuth
// @Router       /users/{id} [get]
```

- [ ] **Step 3: Regenerar + build**

```bash
swag init -g cmd/api/main.go -o docs/swagger --parseInternal --parseDependency
go build ./...
```

Expected: sem erro; `docs/swagger/swagger.json` agora contém `/auth/login`, `/users`, `/users/{id}`.

- [ ] **Step 4: Commit**

```bash
git add cmd/internal/handlers/auth.go cmd/internal/handlers/users.go docs/swagger/
git commit -m "docs: anotacoes swagger de auth e users"
```

---

### Task 6: Anotar handlers de accounts e extrato

**Files:**
- Modify: `cmd/internal/handlers/accounts.go`
- Modify: `cmd/internal/handlers/ledger_entries.go`
- Regenerate: `docs/swagger/`

- [ ] **Step 1: Anotar `CreateAccount`, `GetByID` e `List`**

Em `cmd/internal/handlers/accounts.go`, acima de `func (h *AccountHandler) CreateAccount(`:

```go
// CreateAccount godoc
// @Summary      Cria uma conta pro usuário autenticado
// @Description  O dono da conta vem do token — o body não carrega user_id.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreateAccountRequest  true  "Dados da conta"
// @Success      201   {object}  models.CreateAccountResponse
// @Failure      400,401,500  {object}  apperrors.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts [post]
```

Acima de `func (h *AccountHandler) GetByID(`:

```go
// GetByID godoc
// @Summary      Busca uma conta por ID (apenas do dono)
// @Description  404 se não existe; 403 se a conta é de outro usuário.
// @Tags         accounts
// @Produce      json
// @Param        id  path  string  true  "ID da conta (UUID)"
// @Success      200  {object}  models.Account
// @Failure      400,401,403,404,500  {object}  apperrors.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/{id} [get]
```

Acima de `func (h *AccountHandler) List(`:

```go
// List godoc
// @Summary      Lista as contas do usuário autenticado
// @Tags         accounts
// @Produce      json
// @Success      200  {array}   models.Account
// @Failure      401,500  {object}  apperrors.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts [get]
```

- [ ] **Step 2: Anotar `GetStatement`**

Em `cmd/internal/handlers/ledger_entries.go`, acima de `func (h *LedgerEntryHandler) GetStatement(`:

```go
// GetStatement godoc
// @Summary      Extrato (ledger) de uma conta
// @Description  Lançamentos de débito/crédito da conta; a conta deve pertencer ao usuário do token.
// @Tags         accounts
// @Produce      json
// @Param        id  path  string  true  "ID da conta (UUID)"
// @Success      200  {array}   models.LedgerEntry
// @Failure      400,401,403,404,500  {object}  apperrors.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/{id}/statement [get]
```

- [ ] **Step 3: Regenerar + build**

```bash
swag init -g cmd/api/main.go -o docs/swagger --parseInternal --parseDependency
go build ./...
```

Expected: sem erro; `swagger.json` ganha `/accounts`, `/accounts/{id}`, `/accounts/{id}/statement`.

- [ ] **Step 4: Commit**

```bash
git add cmd/internal/handlers/accounts.go cmd/internal/handlers/ledger_entries.go docs/swagger/
git commit -m "docs: anotacoes swagger de accounts e extrato"
```

---

### Task 7: Anotar handlers de transfers

**Files:**
- Modify: `cmd/internal/handlers/transfers.go`
- Regenerate: `docs/swagger/`

- [ ] **Step 1: Anotar `CreateTransfer` e `GetByID`**

Em `cmd/internal/handlers/transfers.go`, acima de `func (h *TransferHandler) CreateTransfer(`:

```go
// CreateTransfer godoc
// @Summary      Cria uma transferência entre contas
// @Description  A conta de origem deve pertencer ao usuário do token. Valor em centavos. 422 INSUFFICIENT_FUNDS quando o saldo não cobre o valor.
// @Tags         transfers
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreateTransferRequest  true  "Dados da transferência"
// @Success      201   {object}  models.Transfer
// @Failure      400,401,403,404,422,500  {object}  apperrors.ErrorResponse
// @Security     BearerAuth
// @Router       /transfers [post]
```

Acima de `func (h *TransferHandler) GetByID(`:

```go
// GetByID godoc
// @Summary      Busca uma transferência por ID
// @Description  Visível pro dono da conta de origem ou da conta de destino.
// @Tags         transfers
// @Produce      json
// @Param        id  path  string  true  "ID da transferência (UUID)"
// @Success      200  {object}  models.Transfer
// @Failure      400,401,403,404,500  {object}  apperrors.ErrorResponse
// @Security     BearerAuth
// @Router       /transfers/{id} [get]
```

- [ ] **Step 2: Regenerar + build**

```bash
swag init -g cmd/api/main.go -o docs/swagger --parseInternal --parseDependency
go build ./...
```

Expected: sem erro; `swagger.json` ganha `/transfers` e `/transfers/{id}` (com o 422 listado).

- [ ] **Step 3: Commit**

```bash
git add cmd/internal/handlers/transfers.go docs/swagger/
git commit -m "docs: anotacoes swagger de transfers"
```

---

### Task 8: Anotar handlers de participants e pix-keys

**Files:**
- Modify: `cmd/internal/handlers/participants.go`
- Modify: `cmd/internal/handlers/pix_keys.go`
- Regenerate: `docs/swagger/`

- [ ] **Step 1: Anotar `CreateParticipant` e `GetByID`**

Em `cmd/internal/handlers/participants.go`, acima de `func (h *ParticipantHandler) CreateParticipant(`:

```go
// CreateParticipant godoc
// @Summary      Cadastra um participante (instituição financeira)
// @Tags         participants
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreateParticipantRequest  true  "Dados do participante"
// @Success      201   {object}  models.CreateParticipantResponse
// @Failure      400,401,500  {object}  apperrors.ErrorResponse
// @Security     BearerAuth
// @Router       /participants [post]
```

Acima de `func (h *ParticipantHandler) GetByID(`:

```go
// GetByID godoc
// @Summary      Busca um participante por ID
// @Tags         participants
// @Produce      json
// @Param        id  path  string  true  "ID do participante (UUID)"
// @Success      200  {object}  models.Participant
// @Failure      400,401,404,500  {object}  apperrors.ErrorResponse
// @Security     BearerAuth
// @Router       /participants/{id} [get]
```

- [ ] **Step 2: Anotar `CreatePixKey` e `GetByKey`**

Em `cmd/internal/handlers/pix_keys.go`, acima de `func (h *PixKeyHandler) CreatePixKey(`:

```go
// CreatePixKey godoc
// @Summary      Cadastra uma chave PIX pra uma conta
// @Tags         pix-keys
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreatePixKeysRequest  true  "Dados da chave"
// @Success      201   {object}  models.CreatePixKeysResponse
// @Failure      400,401,500  {object}  apperrors.ErrorResponse
// @Security     BearerAuth
// @Router       /pix-keys [post]
```

Acima de `func (h *PixKeyHandler) GetByKey(`:

```go
// GetByKey godoc
// @Summary      Resolve uma chave PIX
// @Tags         pix-keys
// @Produce      json
// @Param        key  path  string  true  "Valor da chave (e-mail, CPF, telefone ou EVP)"
// @Success      200  {object}  models.PixKey
// @Failure      401,404,500  {object}  apperrors.ErrorResponse
// @Security     BearerAuth
// @Router       /pix-keys/{key} [get]
```

- [ ] **Step 3: Regenerar + build**

```bash
swag init -g cmd/api/main.go -o docs/swagger --parseInternal --parseDependency
go build ./...
```

Expected: sem erro; **todas as 13 operações** presentes no `swagger.json`: `/auth/login`, `/users` (post), `/users/{id}`, `/accounts` (get+post), `/accounts/{id}`, `/accounts/{id}/statement`, `/transfers` (post), `/transfers/{id}`, `/participants` (post), `/participants/{id}`, `/pix-keys` (post), `/pix-keys/{key}`.

- [ ] **Step 4: Commit**

```bash
git add cmd/internal/handlers/participants.go cmd/internal/handlers/pix_keys.go docs/swagger/
git commit -m "docs: anotacoes swagger de participants e pix-keys"
```

---

### Task 9: Example tags nos models

**Files:**
- Modify: `cmd/internal/models/users.go`
- Modify: `cmd/internal/models/accounts.go`
- Modify: `cmd/internal/models/transfers.go`
- Modify: `cmd/internal/models/participants.go`
- Modify: `cmd/internal/models/pix_keys.go`
- Regenerate: `docs/swagger/`

(`models/auth.go` já ganhou tags na Task 3; `LedgerEntry` fica sem — é só response e os campos são autoexplicativos.)

- [ ] **Step 1: users.go**

Trocar `CreateUserRequest` (versão pós-Semana 3, campo `password`) por:

```go
type CreateUserRequest struct {
	Name     string `json:"name" example:"Fulano da Silva"`
	Email    string `json:"email" example:"fulano@email.com"`
	TaxID    string `json:"tax_id" example:"12345678901"`
	Password string `json:"password" example:"senha123"`
}
```

- [ ] **Step 2: accounts.go**

Trocar `CreateAccountRequest` (versão pós-Semana 3, sem `UserID`) por:

```go
// CreateAccountRequest é o body do POST /accounts.
// O dono da conta vem do token (context), nunca do body.
type CreateAccountRequest struct {
	ParticipantID uuid.UUID `json:"participantID"`
	AccountNumber string    `json:"accountNumber" example:"12345-6"`
	Branch        string    `json:"branch" example:"0001"`
	AccountType   string    `json:"accountType" enums:"cacc,svgs" example:"cacc"`
}
```

E no struct `Account`, trocar a linha do `Balance` por:

```go
	Balance       int64     `json:"balance" example:"100000"` // centavos
```

- [ ] **Step 3: transfers.go**

Trocar `CreateTransferRequest` por:

```go
type CreateTransferRequest struct {
	SenderAccountID   uuid.UUID `json:"senderAccountID"`
	ReceiverAccountID uuid.UUID `json:"receiverAccountID"`
	Amount            int64     `json:"amount" example:"1000"` // centavos
	Description       string    `json:"description" example:"pagamento do almoço"`
}
```

- [ ] **Step 4: participants.go**

Trocar `CreateParticipantRequest` por:

```go
type CreateParticipantRequest struct {
	Ispb string `json:"ispb" example:"12345678"`
	Name string `json:"name" example:"Banco Dev"`
}
```

- [ ] **Step 5: pix_keys.go**

Trocar `CreatePixKeysRequest` por:

```go
type CreatePixKeysRequest struct {
	KeyType   string    `json:"keyType" enums:"cpf,email,phone,evp" example:"email"`
	KeyValue  string    `json:"keyValue" example:"fulano@email.com"`
	AccountID uuid.UUID `json:"accountID"`
}
```

- [ ] **Step 6: Regenerar + build + testes**

```bash
swag init -g cmd/api/main.go -o docs/swagger --parseInternal --parseDependency
go build ./... ; go test ./...
```

Expected: tudo verde; no `swagger.json`, `CreateAccountRequest.accountType` tem `"enum": ["cacc","svgs"]` e os examples aparecem.

- [ ] **Step 7: Commit**

```bash
git add cmd/internal/models/ docs/swagger/
git commit -m "docs: example e enum tags nos models pro swagger"
```

---

### Task 10: Arquivos `.http` em `requests/`

Convenções: `@baseUrl` no topo; login como named request (`# @name login`) **em cada arquivo** (request variables são file-scoped no REST Client); cada arquivo cobre caminho feliz + ≥1 caso de erro. Sintaxe de referência: `{{login.response.body.$.token}}` (JSONPath com `$.` — formato do REST Client do VS Code).

**Files:**
- Create: `requests/auth.http`
- Create: `requests/users.http`
- Create: `requests/accounts.http`
- Create: `requests/transfers.http`
- Create: `requests/participants.http`
- Create: `requests/pix_keys.http`
- Create: `requests/README.md`

- [ ] **Step 1: `requests/auth.http`**

```http
@baseUrl = http://localhost:8080

### Login (pré-requisito: usuário criado em users.http)
# @name login
POST {{baseUrl}}/auth/login
Content-Type: application/json

{
  "email": "fulano@email.com",
  "password": "senha123"
}

### Erro: senha errada → 401 {"code":"UNAUTHORIZED",...}
POST {{baseUrl}}/auth/login
Content-Type: application/json

{
  "email": "fulano@email.com",
  "password": "senha-errada"
}
```

- [ ] **Step 2: `requests/users.http`**

```http
@baseUrl = http://localhost:8080

### Criar usuário (rode 1x — repetir falha por e-mail duplicado)
# @name createUser
POST {{baseUrl}}/users
Content-Type: application/json

{
  "name": "Fulano da Silva",
  "email": "fulano@email.com",
  "tax_id": "12345678901",
  "password": "senha123"
}

### Login
# @name login
POST {{baseUrl}}/auth/login
Content-Type: application/json

{ "email": "fulano@email.com", "password": "senha123" }

### Buscar o próprio usuário
GET {{baseUrl}}/users/{{createUser.response.body.$.id}}
Authorization: Bearer {{login.response.body.$.token}}

### Erro: sem token → 401 {"code":"UNAUTHORIZED",...}
GET {{baseUrl}}/users/{{createUser.response.body.$.id}}
```

- [ ] **Step 3: `requests/accounts.http`**

```http
@baseUrl = http://localhost:8080

### Login (pré-requisito: users.http rodado)
# @name login
POST {{baseUrl}}/auth/login
Content-Type: application/json

{ "email": "fulano@email.com", "password": "senha123" }

### Participante pro vínculo das contas
# @name createParticipant
POST {{baseUrl}}/participants
Authorization: Bearer {{login.response.body.$.token}}
Content-Type: application/json

{ "ispb": "00000001", "name": "Banco Requests" }

### Criar conta 1
# @name createAccount
POST {{baseUrl}}/accounts
Authorization: Bearer {{login.response.body.$.token}}
Content-Type: application/json

{
  "participantID": "{{createParticipant.response.body.$.id}}",
  "accountNumber": "12345-6",
  "branch": "0001",
  "accountType": "cacc"
}

### Criar conta 2 (destino das transferências em transfers.http)
POST {{baseUrl}}/accounts
Authorization: Bearer {{login.response.body.$.token}}
Content-Type: application/json

{
  "participantID": "{{createParticipant.response.body.$.id}}",
  "accountNumber": "65432-1",
  "branch": "0001",
  "accountType": "svgs"
}

### Buscar conta por ID
GET {{baseUrl}}/accounts/{{createAccount.response.body.$.id}}
Authorization: Bearer {{login.response.body.$.token}}

### Listar minhas contas
GET {{baseUrl}}/accounts
Authorization: Bearer {{login.response.body.$.token}}

### Extrato da conta
GET {{baseUrl}}/accounts/{{createAccount.response.body.$.id}}/statement
Authorization: Bearer {{login.response.body.$.token}}

### Erro: conta inexistente → 404 {"code":"NOT_FOUND",...}
GET {{baseUrl}}/accounts/00000000-0000-0000-0000-00000000dead
Authorization: Bearer {{login.response.body.$.token}}
```

- [ ] **Step 4: `requests/transfers.http`**

```http
@baseUrl = http://localhost:8080

### Login (pré-requisitos: accounts.http rodado — 2 contas — e saldo creditado; ver requests/README.md)
# @name login
POST {{baseUrl}}/auth/login
Content-Type: application/json

{ "email": "fulano@email.com", "password": "senha123" }

### Minhas contas (origem = primeira, destino = segunda)
# @name accounts
GET {{baseUrl}}/accounts
Authorization: Bearer {{login.response.body.$.token}}

### Transferência (caminho feliz — R$ 1,00)
# @name createTransfer
POST {{baseUrl}}/transfers
Authorization: Bearer {{login.response.body.$.token}}
Content-Type: application/json

{
  "senderAccountID": "{{accounts.response.body.$[0].id}}",
  "receiverAccountID": "{{accounts.response.body.$[1].id}}",
  "amount": 100,
  "description": "transferencia de teste"
}

### Buscar transferência por ID
GET {{baseUrl}}/transfers/{{createTransfer.response.body.$.id}}
Authorization: Bearer {{login.response.body.$.token}}

### Erro: saldo insuficiente → 422 {"code":"INSUFFICIENT_FUNDS",...}
POST {{baseUrl}}/transfers
Authorization: Bearer {{login.response.body.$.token}}
Content-Type: application/json

{
  "senderAccountID": "{{accounts.response.body.$[0].id}}",
  "receiverAccountID": "{{accounts.response.body.$[1].id}}",
  "amount": 999999999999,
  "description": "deve falhar"
}
```

- [ ] **Step 5: `requests/participants.http`**

```http
@baseUrl = http://localhost:8080

### Login (pré-requisito: users.http rodado)
# @name login
POST {{baseUrl}}/auth/login
Content-Type: application/json

{ "email": "fulano@email.com", "password": "senha123" }

### Cadastrar participante
# @name createParticipant
POST {{baseUrl}}/participants
Authorization: Bearer {{login.response.body.$.token}}
Content-Type: application/json

{ "ispb": "00000002", "name": "Banco Participants" }

### Buscar participante por ID
GET {{baseUrl}}/participants/{{createParticipant.response.body.$.id}}
Authorization: Bearer {{login.response.body.$.token}}

### Erro: sem ispb → 400 {"code":"INVALID_INPUT",...}
POST {{baseUrl}}/participants
Authorization: Bearer {{login.response.body.$.token}}
Content-Type: application/json

{ "name": "Banco Sem ISPB" }
```

- [ ] **Step 6: `requests/pix_keys.http`**

```http
@baseUrl = http://localhost:8080

### Login (pré-requisitos: users.http e accounts.http rodados)
# @name login
POST {{baseUrl}}/auth/login
Content-Type: application/json

{ "email": "fulano@email.com", "password": "senha123" }

### Minhas contas (a chave vai na primeira)
# @name accounts
GET {{baseUrl}}/accounts
Authorization: Bearer {{login.response.body.$.token}}

### Cadastrar chave PIX (rode 1x — chave é única)
POST {{baseUrl}}/pix-keys
Authorization: Bearer {{login.response.body.$.token}}
Content-Type: application/json

{
  "keyType": "email",
  "keyValue": "fulano@email.com",
  "accountID": "{{accounts.response.body.$[0].id}}"
}

### Resolver chave PIX
GET {{baseUrl}}/pix-keys/fulano@email.com
Authorization: Bearer {{login.response.body.$.token}}

### Erro: chave inexistente → 404 {"code":"NOT_FOUND",...}
GET {{baseUrl}}/pix-keys/nao-existe@x.com
Authorization: Bearer {{login.response.body.$.token}}
```

- [ ] **Step 7: `requests/README.md`**

```markdown
# Requests da API

Arquivos `.http` executáveis contra a API local (`http://localhost:8080`).

## Como usar

- **VS Code:** extensão [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) — clique em "Send Request" acima de cada bloco.
- **GoLand/IntelliJ:** o HTTP Client nativo abre os arquivos, mas a sintaxe de
  request variables (`{{login.response.body.$.token}}`) é do REST Client.
  No JetBrains, troque por response handler (`> {% client.global.set("token",
  response.body.token); %}` no login e `{{token}}` nos demais) ou cole o token
  manualmente.

Cada arquivo é autossuficiente: tem o próprio bloco de login (request variables
não atravessam arquivos) e cobre o caminho feliz + pelo menos um caso de erro,
mostrando o contrato `{"code","message","trace_id"}` na prática.

## Ordem sugerida (banco zerado)

1. `users.http` — cria o usuário e busca por ID.
2. `auth.http` — login (e erro 401).
3. `accounts.http` — participante + 2 contas + extrato.
4. Creditar saldo na conta 1 (não há endpoint de depósito):
   `docker exec -i project_pay_db psql -U pay_user -d project_pay < scripts/dev_seed.sql`
   (ou um `UPDATE accounts SET balance = 100000 WHERE ...` manual).
5. `transfers.http` — transferência + erro 422 de saldo insuficiente.
6. `participants.http` e `pix_keys.http` — em qualquer ordem.
```

- [ ] **Step 8: Verificar de ponta a ponta (precisa da API de pé)**

```bash
docker compose up -d postgres
go run ./cmd/api
```

Executar os arquivos na ordem do `requests/README.md` pelo REST Client. Expected: caminhos felizes 200/201; erros com os status e `code` indicados nos comentários `###`.

- [ ] **Step 9: Commit**

```bash
git add requests/
git commit -m "docs: colecao de arquivos .http com caminho feliz e casos de erro"
```

---

### Task 11: README do projeto + `.env.example`

**Files:**
- Modify: `README.md` (reescrever — preservando a linha de licença)
- Modify: `.env.example`

- [ ] **Step 1: Completar o `.env.example`**

Substituir o conteúdo por (porta 5433 = porta exposta no `docker-compose.yml`; credenciais são as de dev do compose, não segredos):

```bash
# Database (docker-compose expõe o Postgres em localhost:5433)
DB_HOST=localhost
DB_PORT=5433
DB_USER=pay_user
DB_PASSWORD=pay_pass
DB_NAME=project_pay
DB_SSLMODE=disable

# API
API_PORT=8080

# Auth (troque em qualquer ambiente que não seja dev local)
JWT_SECRET=dev-secret-troque-isto
```

- [ ] **Step 2: Reescrever o `README.md`**

Substituir o conteúdo por:

````markdown
# Project Pay

Motor de pagamentos instantâneos inspirado no PIX. Foco em throughput, idempotência e consistência distribuída.

**Fase atual:** Fase 1 — fundação: CRUD das entidades, transferências com controle de saldo, autenticação JWT, frontend mínimo e documentação da API.

## Stack

- **Backend:** Go (stdlib `net/http`), Postgres 16 via [pgx](https://github.com/jackc/pgx) (`pgxpool`)
- **Frontend:** Next.js 15 (App Router) + TypeScript + Tailwind — em `web/`
- **Infra local:** Docker Compose (Postgres + pgAdmin), migrations com [golang-migrate](https://github.com/golang-migrate/migrate)

## Como rodar

Pré-requisitos: Go 1.25+, Docker, Node 20+, CLI do golang-migrate
(`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`).

```bash
# 1. Clonar e configurar
git clone https://github.com/henriquerrodrigues/project_pay.git
cd project_pay
cp .env.example .env          # valores default já funcionam pro compose local

# 2. Banco
docker compose up -d postgres
migrate -path migrations -database "postgres://pay_user:pay_pass@localhost:5433/project_pay?sslmode=disable" up

# 3. (opcional) seed de dev — participante + crédito de saldo pra testar transferência
docker exec -i project_pay_db psql -U pay_user -d project_pay < scripts/dev_seed.sql

# 4. API (porta 8080)
go run ./cmd/api

# 5. Frontend (porta 3000) — em outro terminal
cd web
npm install
npm run dev

# 6. Testes
go test ./...
```

## Arquitetura

```mermaid
flowchart LR
    B[Browser] --> N["Next.js (web/)<br/>proxy + cookie httpOnly"]
    N --> API["API Go :8080"]
    API --> H
    subgraph camadas [camadas da API]
        H[Handler] --> S[Service] --> R[Repository]
    end
    R --> PG[(Postgres 16)]
```

- O browser nunca vê o JWT: o Next guarda o token em cookie httpOnly e faz proxy pra API.
- Handler cuida de HTTP puro; Service concentra regra de negócio, ownership e erros; Repository é SQL puro (pgx).
- Todo erro da API segue o contrato `{"code", "message", "trace_id"}`.
- Modelo de dados: [docs/diagrama.jpg](docs/diagrama.jpg) · Domínio: [docs/dominio.md](docs/dominio.md)

## Documentação da API

- **Swagger UI:** com a API de pé, [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) — use o botão **Authorize** com `Bearer <token>` do `POST /auth/login`.
- **Coleção de requests:** [requests/](requests/) — arquivos `.http` executáveis (REST Client / GoLand), com caminho feliz e casos de erro.

## Estrutura de pastas

```
cmd/
  api/            # main da API (wiring + rotas)
  internal/
    apperrors/    # contrato de erro {"code","message","trace_id"}
    handlers/     # camada HTTP
    services/     # regra de negócio + ownership (um pacote por entidade)
    repositories/ # SQL puro via pgx (um pacote por entidade)
    models/       # structs de request/response e entidades
    middleware/   # auth JWT + request logger (trace ID)
    traceid/      # trace ID no context
docs/
  adr/            # decisões arquiteturais (ADRs)
  swagger/        # pacote gerado pelo swag init (commitado)
migrations/       # migrations SQL versionadas (golang-migrate)
requests/         # coleção .http da API
scripts/          # seed de dev
web/              # frontend Next.js
```

## Decisões arquiteturais

Registradas como ADRs em [docs/adr/](docs/adr/).

## Próximos passos

- **Semana 6:** testes de carga (baseline de throughput).
- Fases seguintes: idempotência de ponta a ponta, concorrência real no processamento e observabilidade.

---

Licenciado sob os termos da [MIT License](LICENSE).
````

- [ ] **Step 3: Conferir a renderização**

Abrir o preview do markdown (VS Code `Ctrl+Shift+V`) — o bloco Mermaid deve renderizar; conferir que a linha de licença está presente.

- [ ] **Step 4: Commit**

```bash
git add README.md .env.example
git commit -m "docs: README completo com como rodar, arquitetura e links da API"
```

---

### Task 12: Primeira ADR

**Files:**
- Create: `docs/adr/README.md`
- Create: `docs/adr/0001-pgx-em-vez-de-database-sql.md`

- [ ] **Step 1: `docs/adr/README.md`**

```markdown
# ADRs — Architecture Decision Records

Uma ADR registra **uma** decisão arquitetural: o contexto em que foi tomada,
a decisão em si e as consequências (boas e ruins). O valor está em ler daqui
a meses e entender *por que* o projeto é do jeito que é — mesmo que a decisão
seja revertida depois.

- Numeração sequencial (`0001-...`, `0002-...`), uma decisão por arquivo.
- ADR não se edita pra "mudar de ideia": escreve-se uma nova que substitui a
  antiga (a antiga vira **Substituída**).

## Índice

| # | Título | Status |
|---|--------|--------|
| [0001](0001-pgx-em-vez-de-database-sql.md) | pgx em vez de database/sql | Aceita |

## Template

    # NNNN — Título da decisão

    ## Status

    Aceita | Substituída por [NNNN](link) | Deprecada

    ## Contexto

    Que problema/força motivou a decisão. Alternativas consideradas.

    ## Decisão

    O que foi decidido, em voz ativa ("Usaremos X").

    ## Consequências

    ### Positivas
    - ...

    ### Negativas
    - ...
```

- [ ] **Step 2: `docs/adr/0001-pgx-em-vez-de-database-sql.md`**

```markdown
# 0001 — pgx em vez de database/sql

## Status

Aceita.

Registro retroativo: a decisão foi tomada na Semana 1 do projeto (junho/2026),
quando o acesso a banco foi implementado; a ADR formaliza o racional depois.

## Contexto

A API precisa acessar Postgres a partir de Go. As alternativas consideradas:

1. **`database/sql` + driver** (`lib/pq` ou `pgx/stdlib`): interface padrão da
   stdlib, agnóstica de banco, ecossistema amplo (sqlx, mocks prontos).
2. **`pgx` nativo** (API própria + `pgxpool`): driver Postgres específico, sem
   passar pela abstração de `database/sql`.

O projeto é um motor de pagamentos com foco em throughput e consistência:
uso intenso de tipos Postgres (`uuid`, `numeric`, `timestamptz`), transações
com `SELECT ... FOR UPDATE` e, nas fases seguintes, batch e `COPY`.
Portabilidade de banco não é requisito — Postgres é decisão de projeto.

## Decisão

Usaremos **pgx nativo (v5)** com pool de conexões via `pgxpool`.

## Consequências

### Positivas

- Suporte de primeira classe aos tipos Postgres (uuid, numeric, timestamptz,
  arrays) sem conversões manuais nem `driver.Valuer`.
- Performance: protocolo binário e prepared statements automáticos por conexão.
- `pgxpool` dá controle fino do pool (tamanho, lifetime, health check).
- Acesso a recursos exclusivos do Postgres quando precisarmos (COPY, LISTEN/NOTIFY,
  batch explícito) — relevantes pras fases de throughput.

### Negativas

- Acoplamento ao Postgres: trocar de banco exigiria reescrever a camada de
  repositories (aceito — Postgres é decisão de projeto).
- API diferente de `database/sql`: menos exemplos genéricos na internet e
  bibliotecas do ecossistema `database/sql` (sqlx, sqlmock) não se aplicam.
- Onboarding: quem só conhece `database/sql` precisa aprender a API do pgx.
```

- [ ] **Step 3: Conferir links**

No preview do markdown, o link do índice (`0001-...`) abre a ADR; o link `docs/adr/` do README do projeto abre este índice.

- [ ] **Step 4: Commit**

```bash
git add docs/adr/
git commit -m "docs: primeira ADR (pgx em vez de database/sql) + template"
```

---

### Task 13: Verificação final (dia 35 — revisão)

Checklist de aceite da semana. Precisa de Postgres + API + front de pé.

- [ ] **Step 1: Geração limpa e build verde**

```bash
swag init -g cmd/api/main.go -o docs/swagger --parseInternal --parseDependency
git status --short docs/swagger/
go build ./... ; go vet ./... ; go test ./...
```

Expected: `swag init` sem warning relevante; `git status` **sem diff** em `docs/swagger/` (gerado está em sincronia com as anotações); build, vet e testes verdes.

- [ ] **Step 2: Smoke da Swagger UI**

```bash
docker compose up -d postgres
go run ./cmd/api
```

No browser, `http://localhost:8080/swagger/index.html`:

1. UI abre com as 6 tags (auth, users, accounts, transfers, participants, pix-keys) e 13 operações.
2. Executar `POST /auth/login` pela UI com um usuário existente → 200 com token.
3. Botão **Authorize** → colar `Bearer <token>` → cadeado fecha nas rotas protegidas.
4. Executar `GET /accounts` → 200 com as contas do usuário.
5. Executar `POST /transfers` (contas e saldo do seed) → 201.
6. Executar `POST /transfers` com `amount` absurdo → 422 com `{"code":"INSUFFICIENT_FUNDS","message":...,"trace_id":...}` — o schema de erro da UI bate com o response real.

- [ ] **Step 3: Arquivos `.http` de ponta a ponta**

Com banco zerado:

```bash
docker compose down -v
docker compose up -d postgres
migrate -path migrations -database "postgres://pay_user:pay_pass@localhost:5433/project_pay?sslmode=disable" up
```

Executar os 6 arquivos na ordem do `requests/README.md`. Expected: todos os blocos com o status esperado (comentários `###` de cada bloco).

- [ ] **Step 4: README "de fora"**

Seguir o "Como rodar" do README literalmente, de cabeça limpa (ideal: outra máquina ou pasta recém-clonada). Anotar qualquer passo que exigiu conhecimento não escrito e corrigir o README na hora.

- [ ] **Step 5: Commit de ajustes (se houver)**

```bash
git add -A
git commit -m "docs: ajustes da revisao final da semana 5"
```

---

## Critério de aceite da Semana 5

- [ ] `/swagger/index.html` abre, Authorize funciona, `POST /transfers` executável pela UI.
- [ ] 13 operações documentadas, todas com `@Failure` apontando pra `apperrors.ErrorResponse`; rotas protegidas com `@Security BearerAuth`.
- [ ] `docs/swagger/` commitado e em sincronia (`swag init` não gera diff).
- [ ] 6 arquivos `.http` + `requests/README.md`, cada um executável de ponta a ponta com ≥1 caso de erro.
- [ ] README com visão, stack, como rodar, Mermaid, links pro Swagger/`requests/`/ADRs, estrutura de pastas, próximos passos e linha de licença.
- [ ] `docs/adr/README.md` (índice + template) e ADR 0001 escrita como registro retroativo.
- [ ] `go build ./... && go vet ./... && go test ./...` verdes.
