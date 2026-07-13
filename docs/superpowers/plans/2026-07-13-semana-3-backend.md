# Semana 3 — Backend (erros, logs, camadas, ownership) — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Backend Go com contrato de erro JSON padronizado (`apperrors`), logs estruturados com `slog` + trace ID por request, camada de services em todas as entidades e validação de ownership (403).

**Architecture:** Handler (HTTP puro) → Service (regra de negócio, ownership, erros `apperrors`) → Repository (SQL puro). Trace ID vive num pacote próprio `traceid` para evitar ciclo de import entre `middleware` e `apperrors`. Services leem o `user_id` do context (chave exportada pelo middleware `Auth`) e dependem de interfaces pequenas dos repositories, definidas no pacote do service (permite fakes nos testes).

**Tech Stack:** Go 1.25, `net/http` (mux stdlib), `log/slog`, pgx v5, `github.com/google/uuid`, JWT existente (não mexer). Testes com stdlib `testing` + `httptest` (sem dependência nova).

**Referência:** spec `docs/superpowers/specs/2026-07-13-semana-3-4-design.md`.

**Fora de escopo:** revisar JWT/login existente, refresh token, testes de handlers HTTP das entidades, pagination no extrato.

**Adição ao spec (lacuna detectada):** `GET /accounts` (lista contas do usuário logado). A tela `/dashboard` da Semana 4 precisa listar as contas do user e não existe endpoint de listagem. Incluído na Task 6.

---

## Estrutura de arquivos (visão geral)

| Arquivo | Responsabilidade |
|---|---|
| `cmd/internal/traceid/traceid.go` (novo) | Chave tipada de context para trace ID + helpers |
| `cmd/internal/apperrors/apperrors.go` (novo) | Tipo `Error`, construtores, `WriteHTTP` |
| `cmd/internal/middleware/request_logger.go` (novo) | Trace ID + log de request |
| `cmd/internal/middleware/auth.go` (modificar) | Helper `UserIDFromContext`; erros 401 via `apperrors` |
| `cmd/internal/services/<entidade>/` (novos) | Um pacote por entidade: users, accounts, ledger_entries, transfers, participants, pix_keys |
| `cmd/internal/handlers/*.go` (modificar) | Passam a depender do service; todo erro sai via `apperrors.WriteHTTP` |
| `cmd/internal/repositories/accounts/accounts.go` (modificar) | `Create` recebe `userID`; novo `ListByUserID` |
| `cmd/api/main.go` (modificar) | slog + wiring repo → service → handler + `RequestLogger` no mux inteiro |
| Removidos | `cmd/internal/models/errors.go`, `cmd/internal/services/services.go`, `cmd/internal/repositories/repositories.go`, `cmd/internal/handlers/handlers.go` |

Comandos rodam da raiz do repo (`c:\Go\repos\project_pay`).

---

### Task 1: Pacote `traceid`

**Files:**
- Create: `cmd/internal/traceid/traceid.go`
- Test: `cmd/internal/traceid/traceid_test.go`

- [ ] **Step 1: Escrever o teste que falha**

```go
// cmd/internal/traceid/traceid_test.go
package traceid_test

import (
	"context"
	"testing"

	"project_pay/cmd/internal/traceid"
)

func TestRoundTrip(t *testing.T) {
	ctx := traceid.NewContext(context.Background(), "abc-123")
	if got := traceid.FromContext(ctx); got != "abc-123" {
		t.Errorf("FromContext = %q, esperava %q", got, "abc-123")
	}
}

func TestMissing(t *testing.T) {
	if got := traceid.FromContext(context.Background()); got != "" {
		t.Errorf("FromContext em context vazio = %q, esperava vazio", got)
	}
}
```

- [ ] **Step 2: Rodar e ver falhar**

Run: `go test ./cmd/internal/traceid/ -v`
Expected: FAIL (pacote `traceid` não existe / não compila)

- [ ] **Step 3: Implementar**

```go
// cmd/internal/traceid/traceid.go
// Package traceid guarda o trace ID do request no context.
// Vive num pacote próprio para que middleware e apperrors possam
// usá-lo sem ciclo de import.
package traceid

import "context"

type ctxKey struct{}

// NewContext retorna um context carregando o trace ID.
func NewContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext retorna o trace ID do context, ou "" se ausente.
func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}
```

- [ ] **Step 4: Rodar e ver passar**

Run: `go test ./cmd/internal/traceid/ -v`
Expected: PASS (2 testes)

- [ ] **Step 5: Commit**

```bash
git add cmd/internal/traceid/
git commit -m "feat: pacote traceid com chave tipada de context"
```

---

### Task 2: Pacote `apperrors`

**Files:**
- Create: `cmd/internal/apperrors/apperrors.go`
- Test: `cmd/internal/apperrors/apperrors_test.go`
- Delete: `cmd/internal/models/errors.go`

- [ ] **Step 1: Escrever os testes que falham**

```go
// cmd/internal/apperrors/apperrors_test.go
package apperrors_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/traceid"
)

type errBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

func doWrite(t *testing.T, err error) (*httptest.ResponseRecorder, errBody) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req = req.WithContext(traceid.NewContext(req.Context(), "tid-1"))
	rec := httptest.NewRecorder()

	apperrors.WriteHTTP(rec, req, err)

	var body errBody
	if jsonErr := json.Unmarshal(rec.Body.Bytes(), &body); jsonErr != nil {
		t.Fatalf("response não é JSON válido: %v — body: %s", jsonErr, rec.Body.String())
	}
	return rec, body
}

func TestWriteHTTP_ErroConhecido(t *testing.T) {
	rec, body := doWrite(t, apperrors.NotFound("conta não encontrada"))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, esperava 404", rec.Code)
	}
	if body.Code != "NOT_FOUND" {
		t.Errorf("code = %q, esperava NOT_FOUND", body.Code)
	}
	if body.Message != "conta não encontrada" {
		t.Errorf("message = %q", body.Message)
	}
	if body.TraceID != "tid-1" {
		t.Errorf("trace_id = %q, esperava tid-1", body.TraceID)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestWriteHTTP_ErroDesconhecidoNaoVaza(t *testing.T) {
	rec, body := doWrite(t, errors.New("pgx: detalhe interno do banco"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, esperava 500", rec.Code)
	}
	if body.Code != "INTERNAL" {
		t.Errorf("code = %q, esperava INTERNAL", body.Code)
	}
	if strings.Contains(rec.Body.String(), "pgx") {
		t.Errorf("detalhe interno vazou no response: %s", rec.Body.String())
	}
}

func TestWriteHTTP_ErroEmbrulhado(t *testing.T) {
	// *apperrors.Error embrulhado com fmt.Errorf %w ainda deve ser reconhecido
	wrapped := errors.Join(apperrors.Forbidden("sem acesso"))
	rec, body := doWrite(t, wrapped)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, esperava 403", rec.Code)
	}
	if body.Code != "FORBIDDEN" {
		t.Errorf("code = %q, esperava FORBIDDEN", body.Code)
	}
}

func TestConstrutores(t *testing.T) {
	tests := []struct {
		name       string
		err        *apperrors.Error
		wantCode   string
		wantStatus int
	}{
		{"NotFound", apperrors.NotFound("x"), "NOT_FOUND", 404},
		{"InvalidInput", apperrors.InvalidInput("x"), "INVALID_INPUT", 400},
		{"Unauthorized", apperrors.Unauthorized("x"), "UNAUTHORIZED", 401},
		{"Forbidden", apperrors.Forbidden("x"), "FORBIDDEN", 403},
		{"InsufficientFunds", apperrors.InsufficientFunds(), "INSUFFICIENT_FUNDS", 422},
		{"Internal", apperrors.Internal(errors.New("boom")), "INTERNAL", 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("Code = %q, esperava %q", tt.err.Code, tt.wantCode)
			}
			if tt.err.HTTPStatus != tt.wantStatus {
				t.Errorf("HTTPStatus = %d, esperava %d", tt.err.HTTPStatus, tt.wantStatus)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	cause := errors.New("causa raiz")
	err := apperrors.Internal(cause)
	if !errors.Is(err, cause) {
		t.Errorf("errors.Is deveria encontrar a causa embrulhada")
	}
}
```

- [ ] **Step 2: Rodar e ver falhar**

Run: `go test ./cmd/internal/apperrors/ -v`
Expected: FAIL (pacote não existe)

- [ ] **Step 3: Implementar**

```go
// cmd/internal/apperrors/apperrors.go
// Package apperrors define o erro de aplicação que atravessa as camadas
// e o contrato JSON de erro da API: {"code","message","trace_id"}.
package apperrors

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"project_pay/cmd/internal/traceid"
)

// Códigos estáveis de máquina — clientes podem depender deles.
const (
	CodeNotFound          = "NOT_FOUND"
	CodeInvalidInput      = "INVALID_INPUT"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	CodeInternal          = "INTERNAL"
)

type Error struct {
	Code       string // código estável de máquina
	HTTPStatus int    // status HTTP correspondente
	Message    string // mensagem segura para o cliente (pt-BR, sem detalhe interno)
	err        error  // causa interna; nunca serializada no response
}

func (e *Error) Error() string {
	if e.err != nil {
		return e.Code + ": " + e.err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *Error) Unwrap() error { return e.err }

func NotFound(msg string) *Error {
	return &Error{Code: CodeNotFound, HTTPStatus: http.StatusNotFound, Message: msg}
}

func InvalidInput(msg string) *Error {
	return &Error{Code: CodeInvalidInput, HTTPStatus: http.StatusBadRequest, Message: msg}
}

func Unauthorized(msg string) *Error {
	return &Error{Code: CodeUnauthorized, HTTPStatus: http.StatusUnauthorized, Message: msg}
}

func Forbidden(msg string) *Error {
	return &Error{Code: CodeForbidden, HTTPStatus: http.StatusForbidden, Message: msg}
}

func InsufficientFunds() *Error {
	return &Error{Code: CodeInsufficientFunds, HTTPStatus: http.StatusUnprocessableEntity, Message: "saldo insuficiente"}
}

func Internal(err error) *Error {
	return &Error{Code: CodeInternal, HTTPStatus: http.StatusInternalServerError, Message: "erro interno", err: err}
}

type response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

// WriteHTTP escreve err no formato do contrato de erro.
// Erros que não são *Error viram 500 genérico; o detalhe vai só pro log.
func WriteHTTP(w http.ResponseWriter, r *http.Request, err error) {
	tid := traceid.FromContext(r.Context())

	var appErr *Error
	if !errors.As(err, &appErr) {
		appErr = Internal(err)
	}

	if appErr.HTTPStatus >= http.StatusInternalServerError {
		slog.Error("erro interno no request",
			"trace_id", tid,
			"error", appErr.Error(),
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)
	json.NewEncoder(w).Encode(response{Code: appErr.Code, Message: appErr.Message, TraceID: tid})
}
```

- [ ] **Step 4: Rodar e ver passar**

Run: `go test ./cmd/internal/apperrors/ -v`
Expected: PASS (todos)

- [ ] **Step 5: Remover o stub antigo**

```bash
rm cmd/internal/models/errors.go
go build ./...
```
Expected: build limpo (`errorResponse` não era usado em lugar nenhum).

- [ ] **Step 6: Commit**

```bash
git add cmd/internal/apperrors/ cmd/internal/models/errors.go
git commit -m "feat: pacote apperrors com contrato JSON de erro"
```

---

### Task 3: `slog` no main + middleware `RequestLogger`

**Files:**
- Create: `cmd/internal/middleware/request_logger.go`
- Test: `cmd/internal/middleware/request_logger_test.go`
- Modify: `cmd/api/main.go`

- [ ] **Step 1: Escrever o teste que falha**

```go
// cmd/internal/middleware/request_logger_test.go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/traceid"
)

func TestRequestLogger_TraceIDNoContextENoHeader(t *testing.T) {
	var tidNoContext string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tidNoContext = traceid.FromContext(r.Context())
		w.WriteHeader(http.StatusTeapot)
	})

	rec := httptest.NewRecorder()
	middleware.RequestLogger(inner).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	tidHeader := rec.Header().Get("X-Trace-Id")
	if tidHeader == "" {
		t.Fatal("header X-Trace-Id ausente")
	}
	if _, err := uuid.Parse(tidHeader); err != nil {
		t.Errorf("X-Trace-Id não é UUID: %q", tidHeader)
	}
	if tidNoContext != tidHeader {
		t.Errorf("trace ID do context (%q) difere do header (%q)", tidNoContext, tidHeader)
	}
	if rec.Code != http.StatusTeapot {
		t.Errorf("status não repassado: %d", rec.Code)
	}
}

func TestRequestLogger_TraceIDsUnicosPorRequest(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := middleware.RequestLogger(inner)

	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/a", nil))
	handler.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/b", nil))

	if rec1.Header().Get("X-Trace-Id") == rec2.Header().Get("X-Trace-Id") {
		t.Error("trace IDs repetidos entre requests")
	}
}
```

- [ ] **Step 2: Rodar e ver falhar**

Run: `go test ./cmd/internal/middleware/ -v`
Expected: FAIL (`RequestLogger` não definido)

- [ ] **Step 3: Implementar o middleware**

```go
// cmd/internal/middleware/request_logger.go
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"project_pay/cmd/internal/traceid"
)

// statusRecorder captura o status code escrito pelo handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// RequestLogger gera um trace ID (UUID v4) por request, injeta no context
// e no header X-Trace-Id, e loga método/rota/status/duração ao final.
// Envolve o mux inteiro.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		tid := uuid.NewString()

		w.Header().Set("X-Trace-Id", tid)
		ctx := traceid.NewContext(r.Context(), tid)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r.WithContext(ctx))

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"trace_id", tid,
		)
	})
}
```

- [ ] **Step 4: Rodar e ver passar**

Run: `go test ./cmd/internal/middleware/ -v`
Expected: PASS (2 testes)

- [ ] **Step 5: Ligar slog e o middleware no main**

Em `cmd/api/main.go`:

1. Adicionar imports `"log/slog"` (e manter `"os"`).
2. Primeira coisa dentro de `main()`, antes de tudo:

```go
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
```

3. Trocar o bloco final (linhas com `log.Printf`/`log.Fatal`) por:

```go
	handler := middleware.RequestLogger(mux)

	port := os.Getenv("API_PORT")
	slog.Info("servidor rodando", "port", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		slog.Error("servidor encerrou com erro", "error", err)
		os.Exit(1)
	}
```

4. Trocar o `log.Fatalf` da conexão com o banco por:

```go
	db, err := database.Connect(context.Background())
	if err != nil {
		slog.Error("erro ao conectar no banco", "error", err)
		os.Exit(1)
	}
```

5. Remover o import `"log"` (não sobra uso).

- [ ] **Step 6: Build + smoke manual**

Run: `go build ./... ; go vet ./...`
Expected: sem erros.

Smoke (opcional, precisa do Postgres do docker-compose de pé): `go run ./cmd/api` e em outro terminal `curl -i http://localhost:8080/users/x` — response deve ter header `X-Trace-Id` e o processo deve logar JSON com `method`, `path`, `status`, `trace_id`.

- [ ] **Step 7: Commit**

```bash
git add cmd/internal/middleware/request_logger.go cmd/internal/middleware/request_logger_test.go cmd/api/main.go
git commit -m "feat: logs estruturados com slog e trace ID por request"
```

---

### Task 4: `middleware.Auth` e handler de login no contrato `apperrors`

**Files:**
- Modify: `cmd/internal/middleware/auth.go`
- Modify: `cmd/internal/handlers/auth.go`

- [ ] **Step 1: Atualizar `middleware/auth.go`**

Substituir o conteúdo por:

```go
// cmd/internal/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	authrepo "project_pay/cmd/internal/repositories/authentication"
)

// contextKey é um tipo próprio para evitar colisão com outras chaves no context
type contextKey string

const UserIDKey contextKey = "user_id"

// UserIDFromContext extrai o user_id injetado pelo middleware Auth.
// Services usam este helper para aplicar ownership.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

// Auth é um middleware que envolve handlers que precisam de autenticação.
// Lê o token do header, valida, extrai o user_id e injeta no context.
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			apperrors.WriteHTTP(w, r, apperrors.Unauthorized("token ausente"))
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		userID, err := authrepo.VerifyToken(tokenString)
		if err != nil {
			apperrors.WriteHTTP(w, r, apperrors.Unauthorized("token inválido"))
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}
```

- [ ] **Step 2: Atualizar `handlers/auth.go`**

Substituir os três `http.Error` do método `Login` por `apperrors`:

```go
// cmd/internal/handlers/auth.go
package handlers

import (
	"encoding/json"
	"net/http"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	authrepo "project_pay/cmd/internal/repositories/authentication"
)

type AuthHandler struct {
	repo *authrepo.Repository
}

func NewAuthHandler(repo *authrepo.Repository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.AuthJWT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("body inválido"))
		return
	}

	// busca o usuário e compara a senha via bcrypt
	userID, ok := h.repo.VerifyCredentials(r.Context(), req.Email, req.Password)
	if !ok {
		apperrors.WriteHTTP(w, r, apperrors.Unauthorized("credenciais inválidas"))
		return
	}

	// gera o token JWT com o user_id nos claims
	token, err := authrepo.CreateToken(userID)
	if err != nil {
		apperrors.WriteHTTP(w, r, apperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
```

- [ ] **Step 3: Build + testes**

Run: `go build ./... ; go test ./...`
Expected: build limpo, testes existentes passam.

- [ ] **Step 4: Commit**

```bash
git add cmd/internal/middleware/auth.go cmd/internal/handlers/auth.go
git commit -m "refactor: auth middleware e login no contrato apperrors"
```

---

### Task 5: Service de users (+ ownership `GET /users/{id}`)

**Files:**
- Create: `cmd/internal/services/users/users.go`
- Test: `cmd/internal/services/users/users_test.go`
- Modify: `cmd/internal/models/users.go` (tag JSON `password_hash` → `password`)
- Modify: `cmd/internal/handlers/users.go`
- Modify: `cmd/api/main.go` (wiring)

- [ ] **Step 1: Corrigir a tag JSON do request**

Em `cmd/internal/models/users.go`, trocar:

```go
type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	TaxID    string `json:"tax_id"`
	Password string `json:"password_hash"`
}
```

por:

```go
type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	TaxID    string `json:"tax_id"`
	Password string `json:"password"`
}
```

E remover o comentário desatualizado acima do struct ("Fase 1: password armazenado direto (sem hash)...") — bcrypt já é aplicado no repository. **Mudança de contrato:** o body do `POST /users` passa a aceitar `password` (era `password_hash`). Nenhum consumidor externo existe; o frontend da Semana 4 já nasce com o nome novo.

- [ ] **Step 2: Escrever os testes do service (falham)**

```go
// cmd/internal/services/users/users_test.go
package users_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"
	userssvc "project_pay/cmd/internal/services/users"
)

type fakeRepo struct {
	users map[uuid.UUID]*models.User
}

func (f *fakeRepo) Create(ctx context.Context, req models.CreateUserRequest) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return f.users[id], nil
}

func ctxWithUser(id uuid.UUID) context.Context {
	return context.WithValue(context.Background(), middleware.UserIDKey, id)
}

func assertAppError(t *testing.T, err error, wantCode string) {
	t.Helper()
	var appErr *apperrors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("esperava *apperrors.Error, veio: %v", err)
	}
	if appErr.Code != wantCode {
		t.Errorf("code = %q, esperava %q", appErr.Code, wantCode)
	}
}

func TestGetByID_ProprioUser(t *testing.T) {
	me := uuid.New()
	svc := userssvc.New(&fakeRepo{users: map[uuid.UUID]*models.User{
		me: {ID: me, Name: "Eu"},
	}})

	user, err := svc.GetByID(ctxWithUser(me), me)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if user.ID != me {
		t.Errorf("retornou user errado: %v", user.ID)
	}
}

func TestGetByID_OutroUser403(t *testing.T) {
	me, other := uuid.New(), uuid.New()
	svc := userssvc.New(&fakeRepo{users: map[uuid.UUID]*models.User{
		other: {ID: other, Name: "Outro"},
	}})

	_, err := svc.GetByID(ctxWithUser(me), other)
	assertAppError(t, err, apperrors.CodeForbidden)
}

func TestGetByID_Inexistente404(t *testing.T) {
	me := uuid.New()
	svc := userssvc.New(&fakeRepo{users: map[uuid.UUID]*models.User{}})

	_, err := svc.GetByID(ctxWithUser(me), uuid.New())
	assertAppError(t, err, apperrors.CodeNotFound)
}

func TestGetByID_SemUserNoContext401(t *testing.T) {
	svc := userssvc.New(&fakeRepo{users: map[uuid.UUID]*models.User{}})

	_, err := svc.GetByID(context.Background(), uuid.New())
	assertAppError(t, err, apperrors.CodeUnauthorized)
}

func TestCreate_CamposObrigatorios(t *testing.T) {
	svc := userssvc.New(&fakeRepo{})

	_, err := svc.Create(context.Background(), models.CreateUserRequest{
		Name: "Fulano", Email: "", Password: "123",
	})
	assertAppError(t, err, apperrors.CodeInvalidInput)
}

func TestCreate_OK(t *testing.T) {
	svc := userssvc.New(&fakeRepo{})

	id, err := svc.Create(context.Background(), models.CreateUserRequest{
		Name: "Fulano", Email: "f@x.com", TaxID: "123", Password: "senha",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if id == uuid.Nil {
		t.Error("id vazio")
	}
}
```

- [ ] **Step 3: Rodar e ver falhar**

Run: `go test ./cmd/internal/services/users/ -v`
Expected: FAIL (pacote não existe)

- [ ] **Step 4: Implementar o service**

```go
// cmd/internal/services/users/users.go
// Package users contém a regra de negócio de usuários.
package users

import (
	"context"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"
)

// UserRepository é o recorte do repository que o service precisa.
// A interface vive aqui para permitir fake nos testes.
type UserRepository interface {
	Create(ctx context.Context, req models.CreateUserRequest) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type Service struct {
	repo UserRepository
}

func New(repo UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req models.CreateUserRequest) (uuid.UUID, error) {
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return uuid.Nil, apperrors.InvalidInput("name, email e password são obrigatórios")
	}

	id, err := s.repo.Create(ctx, req)
	if err != nil {
		return uuid.Nil, apperrors.Internal(err)
	}
	return id, nil
}

// GetByID só permite o próprio usuário: recurso inexistente é 404,
// existente mas de outro user é 403.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	authUserID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if user == nil {
		return nil, apperrors.NotFound("usuário não encontrado")
	}
	if user.ID != authUserID {
		return nil, apperrors.Forbidden("acesso negado")
	}
	return user, nil
}
```

- [ ] **Step 5: Rodar e ver passar**

Run: `go test ./cmd/internal/services/users/ -v`
Expected: PASS (6 testes)

- [ ] **Step 6: Refatorar o handler**

Substituir `cmd/internal/handlers/users.go` por:

```go
// cmd/internal/handlers/users.go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	userssvc "project_pay/cmd/internal/services/users"
)

type UserHandler struct {
	svc *userssvc.Service
}

func NewUserHandler(svc *userssvc.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("body inválido"))
		return
	}

	id, err := h.svc.Create(r.Context(), req)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.CreateUserResponse{ID: id})
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("id inválido"))
		return
	}

	user, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
```

- [ ] **Step 7: Atualizar o wiring no main**

Em `cmd/api/main.go`, adicionar o import:

```go
	userssvc "project_pay/cmd/internal/services/users"
```

e trocar:

```go
	userRepo := usersrepo.New(db)
	userHandler := handlers.NewUserHandler(userRepo)
```

por:

```go
	userRepo := usersrepo.New(db)
	userSvc := userssvc.New(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)
```

- [ ] **Step 8: Build + testes completos**

Run: `go build ./... ; go test ./...`
Expected: tudo verde.

- [ ] **Step 9: Commit**

```bash
git add cmd/internal/services/users/ cmd/internal/handlers/users.go cmd/internal/models/users.go cmd/api/main.go
git commit -m "feat: service de users com ownership no GET /users/{id}"
```

---

### Task 6: Service de accounts (+ vínculo ao token, ownership, `GET /accounts`)

**Files:**
- Modify: `cmd/internal/models/accounts.go` (remove `UserID` do request)
- Modify: `cmd/internal/repositories/accounts/accounts.go` (`Create` recebe userID; novo `ListByUserID`)
- Create: `cmd/internal/services/accounts/accounts.go`
- Test: `cmd/internal/services/accounts/accounts_test.go`
- Modify: `cmd/internal/handlers/accounts.go`
- Modify: `cmd/api/main.go` (wiring + rota `GET /accounts`)

- [ ] **Step 1: Modelo — body não aceita mais `user_id`**

Em `cmd/internal/models/accounts.go`, trocar `CreateAccountRequest` por:

```go
// CreateAccountRequest é o body do POST /accounts.
// O dono da conta vem do token (context), nunca do body.
type CreateAccountRequest struct {
	ParticipantID uuid.UUID `json:"participantID"`
	AccountNumber string    `json:"accountNumber"`
	Branch        string    `json:"branch"`
	AccountType   string    `json:"accountType"`
}
```

(`Account` e `CreateAccountResponse` ficam como estão.)

- [ ] **Step 2: Repository — `Create(ctx, userID, req)` + `ListByUserID`**

Em `cmd/internal/repositories/accounts/accounts.go`, trocar o método `Create` por:

```go
func (r *Repository) Create(ctx context.Context, userID uuid.UUID, req models.CreateAccountRequest) (uuid.UUID, error) {
	var id uuid.UUID

	err := r.db.QueryRow(ctx, sqlCreateAccount,
		userID,
		req.ParticipantID,
		req.AccountNumber,
		req.Branch,
		req.AccountType,
	).Scan(&id)

	return id, err
}
```

e adicionar ao final do arquivo:

```go
func (r *Repository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	rows, err := r.db.Query(ctx, sqlListByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []models.Account{}
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.ParticipantID,
			&a.AccountNumber,
			&a.Branch,
			&a.AccountType,
			&a.Balance,
			&a.CreatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

const sqlListByUserID = `
	SELECT id, user_id, participant_id, account_number, branch, account_type, balance, created_at
	FROM accounts
	WHERE user_id = $1
	ORDER BY created_at`
```

(O SQL `sqlCreateAccount` existente não muda — a ordem dos parâmetros continua `user_id, participant_id, ...`.)

- [ ] **Step 3: Escrever os testes do service (falham)**

```go
// cmd/internal/services/accounts/accounts_test.go
package accounts_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"
	accountssvc "project_pay/cmd/internal/services/accounts"
)

type fakeRepo struct {
	accounts      map[uuid.UUID]*models.Account
	createdUserID uuid.UUID // captura o userID passado no Create
}

func (f *fakeRepo) Create(ctx context.Context, userID uuid.UUID, req models.CreateAccountRequest) (uuid.UUID, error) {
	f.createdUserID = userID
	return uuid.New(), nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	return f.accounts[id], nil
}

func (f *fakeRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	out := []models.Account{}
	for _, a := range f.accounts {
		if a.UserID == userID {
			out = append(out, *a)
		}
	}
	return out, nil
}

func ctxWithUser(id uuid.UUID) context.Context {
	return context.WithValue(context.Background(), middleware.UserIDKey, id)
}

func assertAppError(t *testing.T, err error, wantCode string) {
	t.Helper()
	var appErr *apperrors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("esperava *apperrors.Error, veio: %v", err)
	}
	if appErr.Code != wantCode {
		t.Errorf("code = %q, esperava %q", appErr.Code, wantCode)
	}
}

func validReq() models.CreateAccountRequest {
	return models.CreateAccountRequest{
		ParticipantID: uuid.New(),
		AccountNumber: "12345-6",
		Branch:        "0001",
		AccountType:   "cacc",
	}
}

func TestCreate_VinculaUserDoToken(t *testing.T) {
	me := uuid.New()
	repo := &fakeRepo{}
	svc := accountssvc.New(repo)

	id, err := svc.Create(ctxWithUser(me), validReq())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if id == uuid.Nil {
		t.Error("id vazio")
	}
	if repo.createdUserID != me {
		t.Errorf("conta criada com userID %v, esperava %v (do token)", repo.createdUserID, me)
	}
}

func TestCreate_AccountTypeInvalido(t *testing.T) {
	svc := accountssvc.New(&fakeRepo{})
	req := validReq()
	req.AccountType = "xxxx"

	_, err := svc.Create(ctxWithUser(uuid.New()), req)
	assertAppError(t, err, apperrors.CodeInvalidInput)
}

func TestCreate_CamposObrigatorios(t *testing.T) {
	svc := accountssvc.New(&fakeRepo{})
	req := validReq()
	req.AccountNumber = ""

	_, err := svc.Create(ctxWithUser(uuid.New()), req)
	assertAppError(t, err, apperrors.CodeInvalidInput)
}

func TestGetByID_Dono(t *testing.T) {
	me, accID := uuid.New(), uuid.New()
	svc := accountssvc.New(&fakeRepo{accounts: map[uuid.UUID]*models.Account{
		accID: {ID: accID, UserID: me, Balance: 100},
	}})

	acc, err := svc.GetByID(ctxWithUser(me), accID)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if acc.ID != accID {
		t.Errorf("conta errada: %v", acc.ID)
	}
}

func TestGetByID_OutroDono403(t *testing.T) {
	me, accID := uuid.New(), uuid.New()
	svc := accountssvc.New(&fakeRepo{accounts: map[uuid.UUID]*models.Account{
		accID: {ID: accID, UserID: uuid.New()}, // dona é outra pessoa
	}})

	_, err := svc.GetByID(ctxWithUser(me), accID)
	assertAppError(t, err, apperrors.CodeForbidden)
}

func TestGetByID_Inexistente404(t *testing.T) {
	svc := accountssvc.New(&fakeRepo{accounts: map[uuid.UUID]*models.Account{}})

	_, err := svc.GetByID(ctxWithUser(uuid.New()), uuid.New())
	assertAppError(t, err, apperrors.CodeNotFound)
}

func TestListByCurrentUser_SoMinhasContas(t *testing.T) {
	me := uuid.New()
	minha, deOutro := uuid.New(), uuid.New()
	svc := accountssvc.New(&fakeRepo{accounts: map[uuid.UUID]*models.Account{
		minha:   {ID: minha, UserID: me},
		deOutro: {ID: deOutro, UserID: uuid.New()},
	}})

	accounts, err := svc.ListByCurrentUser(ctxWithUser(me))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(accounts) != 1 || accounts[0].ID != minha {
		t.Errorf("esperava só a minha conta, veio: %v", accounts)
	}
}
```

- [ ] **Step 4: Rodar e ver falhar**

Run: `go test ./cmd/internal/services/accounts/ -v`
Expected: FAIL (pacote não existe)

- [ ] **Step 5: Implementar o service**

```go
// cmd/internal/services/accounts/accounts.go
// Package accounts contém a regra de negócio de contas.
package accounts

import (
	"context"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"
)

// AccountRepository é o recorte do repository que o service precisa.
type AccountRepository interface {
	Create(ctx context.Context, userID uuid.UUID, req models.CreateAccountRequest) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Account, error)
}

type Service struct {
	repo AccountRepository
}

func New(repo AccountRepository) *Service {
	return &Service{repo: repo}
}

// Create cria uma conta vinculada ao user do token — o body não carrega user_id.
func (s *Service) Create(ctx context.Context, req models.CreateAccountRequest) (uuid.UUID, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return uuid.Nil, apperrors.Unauthorized("usuário não autenticado")
	}

	if req.ParticipantID == uuid.Nil || req.AccountNumber == "" || req.Branch == "" {
		return uuid.Nil, apperrors.InvalidInput("participantID, accountNumber e branch são obrigatórios")
	}
	if req.AccountType != "cacc" && req.AccountType != "svgs" {
		return uuid.Nil, apperrors.InvalidInput("accountType deve ser 'cacc' ou 'svgs'")
	}

	id, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return uuid.Nil, apperrors.Internal(err)
	}
	return id, nil
}

// GetByID aplica ownership: inexistente é 404, de outro user é 403.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if account == nil {
		return nil, apperrors.NotFound("conta não encontrada")
	}
	if account.UserID != userID {
		return nil, apperrors.Forbidden("conta não pertence ao usuário")
	}
	return account, nil
}

// ListByCurrentUser lista as contas do user autenticado.
func (s *Service) ListByCurrentUser(ctx context.Context) ([]models.Account, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	accounts, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return accounts, nil
}
```

- [ ] **Step 6: Rodar e ver passar**

Run: `go test ./cmd/internal/services/accounts/ -v`
Expected: PASS (7 testes)

- [ ] **Step 7: Refatorar o handler (+ método `List`)**

Substituir `cmd/internal/handlers/accounts.go` por:

```go
// cmd/internal/handlers/accounts.go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	accountssvc "project_pay/cmd/internal/services/accounts"
)

type AccountHandler struct {
	svc *accountssvc.Service
}

func NewAccountHandler(svc *accountssvc.Service) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("body inválido"))
		return
	}

	id, err := h.svc.Create(r.Context(), req)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.CreateAccountResponse{ID: id})
}

func (h *AccountHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("id inválido"))
		return
	}

	account, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

// List responde GET /accounts com as contas do user autenticado.
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.svc.ListByCurrentUser(r.Context())
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}
```

- [ ] **Step 8: Wiring + rota nova no main**

Em `cmd/api/main.go`, adicionar o import:

```go
	accountssvc "project_pay/cmd/internal/services/accounts"
```

trocar:

```go
	accountRepo := accountsrepo.New(db)
	accountHandler := handlers.NewAccountHandler(accountRepo)
```

por:

```go
	accountRepo := accountsrepo.New(db)
	accountSvc := accountssvc.New(accountRepo)
	accountHandler := handlers.NewAccountHandler(accountSvc)
```

e adicionar a rota (junto das outras de accounts):

```go
	mux.HandleFunc("GET /accounts", middleware.Auth(accountHandler.List))
```

- [ ] **Step 9: Build + testes completos**

Run: `go build ./... ; go test ./...`
Expected: tudo verde.

- [ ] **Step 10: Commit**

```bash
git add cmd/internal/models/accounts.go cmd/internal/repositories/accounts/ cmd/internal/services/accounts/ cmd/internal/handlers/accounts.go cmd/api/main.go
git commit -m "feat: service de accounts com ownership, vinculo ao token e GET /accounts"
```

---

### Task 7: Service do extrato (ledger_entries) com ownership

**Files:**
- Create: `cmd/internal/services/ledger_entries/ledger_entries.go`
- Test: `cmd/internal/services/ledger_entries/ledger_entries_test.go`
- Modify: `cmd/internal/handlers/ledger_entries.go`
- Modify: `cmd/api/main.go` (wiring)

- [ ] **Step 1: Escrever os testes (falham)**

```go
// cmd/internal/services/ledger_entries/ledger_entries_test.go
package ledgerentries_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"
	ledgersvc "project_pay/cmd/internal/services/ledger_entries"
)

type fakeLedgerRepo struct {
	entries map[uuid.UUID][]models.LedgerEntry // accountID -> entries
}

func (f *fakeLedgerRepo) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]models.LedgerEntry, error) {
	return f.entries[accountID], nil
}

type fakeAccountGetter struct {
	accounts map[uuid.UUID]*models.Account
}

func (f *fakeAccountGetter) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	return f.accounts[id], nil
}

func ctxWithUser(id uuid.UUID) context.Context {
	return context.WithValue(context.Background(), middleware.UserIDKey, id)
}

func assertAppError(t *testing.T, err error, wantCode string) {
	t.Helper()
	var appErr *apperrors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("esperava *apperrors.Error, veio: %v", err)
	}
	if appErr.Code != wantCode {
		t.Errorf("code = %q, esperava %q", appErr.Code, wantCode)
	}
}

func TestGetStatement_Dono(t *testing.T) {
	me, accID := uuid.New(), uuid.New()
	svc := ledgersvc.New(
		&fakeLedgerRepo{entries: map[uuid.UUID][]models.LedgerEntry{
			accID: {{ID: uuid.New(), AccountID: accID, Amount: 100}},
		}},
		&fakeAccountGetter{accounts: map[uuid.UUID]*models.Account{
			accID: {ID: accID, UserID: me},
		}},
	)

	entries, err := svc.GetStatement(ctxWithUser(me), accID)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("esperava 1 entry, veio %d", len(entries))
	}
}

func TestGetStatement_OutroDono403(t *testing.T) {
	me, accID := uuid.New(), uuid.New()
	svc := ledgersvc.New(
		&fakeLedgerRepo{entries: map[uuid.UUID][]models.LedgerEntry{}},
		&fakeAccountGetter{accounts: map[uuid.UUID]*models.Account{
			accID: {ID: accID, UserID: uuid.New()},
		}},
	)

	_, err := svc.GetStatement(ctxWithUser(me), accID)
	assertAppError(t, err, apperrors.CodeForbidden)
}

func TestGetStatement_ContaInexistente404(t *testing.T) {
	svc := ledgersvc.New(
		&fakeLedgerRepo{entries: map[uuid.UUID][]models.LedgerEntry{}},
		&fakeAccountGetter{accounts: map[uuid.UUID]*models.Account{}},
	)

	_, err := svc.GetStatement(ctxWithUser(uuid.New()), uuid.New())
	assertAppError(t, err, apperrors.CodeNotFound)
}
```

- [ ] **Step 2: Rodar e ver falhar**

Run: `go test ./cmd/internal/services/ledger_entries/ -v`
Expected: FAIL (pacote não existe)

- [ ] **Step 3: Implementar o service**

```go
// cmd/internal/services/ledger_entries/ledger_entries.go
// Package ledgerentries contém a regra de negócio do extrato.
package ledgerentries

import (
	"context"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"
)

type LedgerRepository interface {
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]models.LedgerEntry, error)
}

// AccountGetter é o recorte do repository de accounts usado pra checar ownership.
type AccountGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
}

type Service struct {
	repo     LedgerRepository
	accounts AccountGetter
}

func New(repo LedgerRepository, accounts AccountGetter) *Service {
	return &Service{repo: repo, accounts: accounts}
}

// GetStatement retorna o extrato; a conta deve pertencer ao user do token.
func (s *Service) GetStatement(ctx context.Context, accountID uuid.UUID) ([]models.LedgerEntry, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if account == nil {
		return nil, apperrors.NotFound("conta não encontrada")
	}
	if account.UserID != userID {
		return nil, apperrors.Forbidden("conta não pertence ao usuário")
	}

	entries, err := s.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return entries, nil
}
```

- [ ] **Step 4: Rodar e ver passar**

Run: `go test ./cmd/internal/services/ledger_entries/ -v`
Expected: PASS (3 testes)

- [ ] **Step 5: Refatorar o handler**

Substituir `cmd/internal/handlers/ledger_entries.go` por:

```go
// cmd/internal/handlers/ledger_entries.go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	ledgersvc "project_pay/cmd/internal/services/ledger_entries"
)

type LedgerEntryHandler struct {
	svc *ledgersvc.Service
}

func NewLedgerEntryHandler(svc *ledgersvc.Service) *LedgerEntryHandler {
	return &LedgerEntryHandler{svc: svc}
}

func (h *LedgerEntryHandler) GetStatement(w http.ResponseWriter, r *http.Request) {
	accountID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("id inválido"))
		return
	}

	entries, err := h.svc.GetStatement(r.Context(), accountID)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
```

- [ ] **Step 6: Wiring no main**

Em `cmd/api/main.go`, adicionar o import:

```go
	ledgersvc "project_pay/cmd/internal/services/ledger_entries"
```

e trocar:

```go
	ledgerRepo := ledgerrepo.New(db)
	ledgerHandler := handlers.NewLedgerEntryHandler(ledgerRepo)
```

por:

```go
	ledgerRepo := ledgerrepo.New(db)
	ledgerSvc := ledgersvc.New(ledgerRepo, accountRepo)
	ledgerHandler := handlers.NewLedgerEntryHandler(ledgerSvc)
```

(`accountRepo` já existe no main — o `*accountsrepo.Repository` satisfaz a interface `AccountGetter`. O bloco de accounts precisa vir antes deste no main.)

- [ ] **Step 7: Build + testes completos**

Run: `go build ./... ; go test ./...`
Expected: tudo verde.

- [ ] **Step 8: Commit**

```bash
git add cmd/internal/services/ledger_entries/ cmd/internal/handlers/ledger_entries.go cmd/api/main.go
git commit -m "feat: service do extrato com ownership da conta"
```

---

### Task 8: Service de transfers (validações + ownership)

**Files:**
- Create: `cmd/internal/services/transfers/transfers.go`
- Test: `cmd/internal/services/transfers/transfers_test.go`
- Modify: `cmd/internal/handlers/transfers.go`
- Modify: `cmd/api/main.go` (wiring)

- [ ] **Step 1: Escrever os testes (falham)**

```go
// cmd/internal/services/transfers/transfers_test.go
package transfers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"
	transfersrepo "project_pay/cmd/internal/repositories/transfers"
	transferssvc "project_pay/cmd/internal/services/transfers"
)

type fakeTransferRepo struct {
	transfers map[uuid.UUID]*models.Transfer
	createErr error // erro a retornar no Create (ex.: ErrInsufficientFunds)
}

func (f *fakeTransferRepo) Create(ctx context.Context, req models.CreateTransferRequest) (*models.Transfer, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &models.Transfer{
		ID:                uuid.New(),
		SenderAccountID:   req.SenderAccountID,
		ReceiverAccountID: req.ReceiverAccountID,
		Amount:            req.Amount,
		Status:            "completed",
	}, nil
}

func (f *fakeTransferRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Transfer, error) {
	return f.transfers[id], nil
}

type fakeAccountGetter struct {
	accounts map[uuid.UUID]*models.Account
}

func (f *fakeAccountGetter) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	return f.accounts[id], nil
}

func ctxWithUser(id uuid.UUID) context.Context {
	return context.WithValue(context.Background(), middleware.UserIDKey, id)
}

func assertAppError(t *testing.T, err error, wantCode string) {
	t.Helper()
	var appErr *apperrors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("esperava *apperrors.Error, veio: %v", err)
	}
	if appErr.Code != wantCode {
		t.Errorf("code = %q, esperava %q", appErr.Code, wantCode)
	}
}

// fixture: user "me" dono da conta origem; conta destino de outra pessoa
func setup() (me uuid.UUID, senderID uuid.UUID, receiverID uuid.UUID, accounts *fakeAccountGetter) {
	me, senderID, receiverID = uuid.New(), uuid.New(), uuid.New()
	accounts = &fakeAccountGetter{accounts: map[uuid.UUID]*models.Account{
		senderID:   {ID: senderID, UserID: me},
		receiverID: {ID: receiverID, UserID: uuid.New()},
	}}
	return me, senderID, receiverID, accounts
}

func TestCreate_OK(t *testing.T) {
	me, senderID, receiverID, accounts := setup()
	svc := transferssvc.New(&fakeTransferRepo{}, accounts)

	transfer, err := svc.Create(ctxWithUser(me), models.CreateTransferRequest{
		SenderAccountID: senderID, ReceiverAccountID: receiverID, Amount: 100,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if transfer.Amount != 100 {
		t.Errorf("amount = %d", transfer.Amount)
	}
}

func TestCreate_ValorInvalido(t *testing.T) {
	me, senderID, receiverID, accounts := setup()
	svc := transferssvc.New(&fakeTransferRepo{}, accounts)

	_, err := svc.Create(ctxWithUser(me), models.CreateTransferRequest{
		SenderAccountID: senderID, ReceiverAccountID: receiverID, Amount: 0,
	})
	assertAppError(t, err, apperrors.CodeInvalidInput)
}

func TestCreate_MesmaContaOrigemDestino(t *testing.T) {
	me, senderID, _, accounts := setup()
	svc := transferssvc.New(&fakeTransferRepo{}, accounts)

	_, err := svc.Create(ctxWithUser(me), models.CreateTransferRequest{
		SenderAccountID: senderID, ReceiverAccountID: senderID, Amount: 100,
	})
	assertAppError(t, err, apperrors.CodeInvalidInput)
}

func TestCreate_OrigemDeOutroUser403(t *testing.T) {
	me, senderID, receiverID, accounts := setup()
	// origem pertence a outra pessoa
	accounts.accounts[senderID].UserID = uuid.New()
	svc := transferssvc.New(&fakeTransferRepo{}, accounts)

	_, err := svc.Create(ctxWithUser(me), models.CreateTransferRequest{
		SenderAccountID: senderID, ReceiverAccountID: receiverID, Amount: 100,
	})
	assertAppError(t, err, apperrors.CodeForbidden)
}

func TestCreate_OrigemInexistente404(t *testing.T) {
	me, _, receiverID, accounts := setup()
	svc := transferssvc.New(&fakeTransferRepo{}, accounts)

	_, err := svc.Create(ctxWithUser(me), models.CreateTransferRequest{
		SenderAccountID: uuid.New(), ReceiverAccountID: receiverID, Amount: 100,
	})
	assertAppError(t, err, apperrors.CodeNotFound)
}

func TestCreate_DestinoInexistente404(t *testing.T) {
	me, senderID, _, accounts := setup()
	svc := transferssvc.New(&fakeTransferRepo{}, accounts)

	_, err := svc.Create(ctxWithUser(me), models.CreateTransferRequest{
		SenderAccountID: senderID, ReceiverAccountID: uuid.New(), Amount: 100,
	})
	assertAppError(t, err, apperrors.CodeNotFound)
}

func TestCreate_SaldoInsuficiente422(t *testing.T) {
	me, senderID, receiverID, accounts := setup()
	svc := transferssvc.New(&fakeTransferRepo{createErr: transfersrepo.ErrInsufficientFunds}, accounts)

	_, err := svc.Create(ctxWithUser(me), models.CreateTransferRequest{
		SenderAccountID: senderID, ReceiverAccountID: receiverID, Amount: 100,
	})
	assertAppError(t, err, apperrors.CodeInsufficientFunds)
}

func TestGetByID_DonoDaOrigem(t *testing.T) {
	me, senderID, receiverID, accounts := setup()
	transferID := uuid.New()
	svc := transferssvc.New(&fakeTransferRepo{transfers: map[uuid.UUID]*models.Transfer{
		transferID: {ID: transferID, SenderAccountID: senderID, ReceiverAccountID: receiverID},
	}}, accounts)

	transfer, err := svc.GetByID(ctxWithUser(me), transferID)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if transfer.ID != transferID {
		t.Errorf("transfer errada: %v", transfer.ID)
	}
}

func TestGetByID_DonoDoDestino(t *testing.T) {
	_, senderID, receiverID, accounts := setup()
	receiverOwner := accounts.accounts[receiverID].UserID
	transferID := uuid.New()
	svc := transferssvc.New(&fakeTransferRepo{transfers: map[uuid.UUID]*models.Transfer{
		transferID: {ID: transferID, SenderAccountID: senderID, ReceiverAccountID: receiverID},
	}}, accounts)

	if _, err := svc.GetByID(ctxWithUser(receiverOwner), transferID); err != nil {
		t.Fatalf("dono do destino deveria ver a transfer: %v", err)
	}
}

func TestGetByID_Terceiro403(t *testing.T) {
	_, senderID, receiverID, accounts := setup()
	transferID := uuid.New()
	svc := transferssvc.New(&fakeTransferRepo{transfers: map[uuid.UUID]*models.Transfer{
		transferID: {ID: transferID, SenderAccountID: senderID, ReceiverAccountID: receiverID},
	}}, accounts)

	_, err := svc.GetByID(ctxWithUser(uuid.New()), transferID)
	assertAppError(t, err, apperrors.CodeForbidden)
}

func TestGetByID_Inexistente404(t *testing.T) {
	me, _, _, accounts := setup()
	svc := transferssvc.New(&fakeTransferRepo{transfers: map[uuid.UUID]*models.Transfer{}}, accounts)

	_, err := svc.GetByID(ctxWithUser(me), uuid.New())
	assertAppError(t, err, apperrors.CodeNotFound)
}
```

- [ ] **Step 2: Rodar e ver falhar**

Run: `go test ./cmd/internal/services/transfers/ -v`
Expected: FAIL (pacote não existe)

- [ ] **Step 3: Implementar o service**

```go
// cmd/internal/services/transfers/transfers.go
// Package transfers contém a regra de negócio de transferências.
package transfers

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"
	transfersrepo "project_pay/cmd/internal/repositories/transfers"
)

type TransferRepository interface {
	Create(ctx context.Context, req models.CreateTransferRequest) (*models.Transfer, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transfer, error)
}

// AccountGetter é o recorte do repository de accounts usado pra checar ownership.
type AccountGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
}

type Service struct {
	repo     TransferRepository
	accounts AccountGetter
}

func New(repo TransferRepository, accounts AccountGetter) *Service {
	return &Service{repo: repo, accounts: accounts}
}

// Create valida o negócio e exige que a conta origem pertença ao user do token.
func (s *Service) Create(ctx context.Context, req models.CreateTransferRequest) (*models.Transfer, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	if req.Amount <= 0 {
		return nil, apperrors.InvalidInput("valor deve ser maior que zero")
	}
	if req.SenderAccountID == req.ReceiverAccountID {
		return nil, apperrors.InvalidInput("conta de origem e destino não podem ser iguais")
	}

	sender, err := s.accounts.GetByID(ctx, req.SenderAccountID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if sender == nil {
		return nil, apperrors.NotFound("conta de origem não encontrada")
	}
	if sender.UserID != userID {
		return nil, apperrors.Forbidden("conta de origem não pertence ao usuário")
	}

	receiver, err := s.accounts.GetByID(ctx, req.ReceiverAccountID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if receiver == nil {
		return nil, apperrors.NotFound("conta de destino não encontrada")
	}

	transfer, err := s.repo.Create(ctx, req)
	if errors.Is(err, transfersrepo.ErrInsufficientFunds) {
		return nil, apperrors.InsufficientFunds()
	}
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return transfer, nil
}

// GetByID permite ver a transfer a quem for dono da conta origem OU destino.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Transfer, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	transfer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if transfer == nil {
		return nil, apperrors.NotFound("transferência não encontrada")
	}

	owns, err := s.userOwnsAccount(ctx, userID, transfer.SenderAccountID)
	if err != nil {
		return nil, err
	}
	if !owns {
		owns, err = s.userOwnsAccount(ctx, userID, transfer.ReceiverAccountID)
		if err != nil {
			return nil, err
		}
	}
	if !owns {
		return nil, apperrors.Forbidden("transferência não pertence ao usuário")
	}
	return transfer, nil
}

func (s *Service) userOwnsAccount(ctx context.Context, userID, accountID uuid.UUID) (bool, error) {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return false, apperrors.Internal(err)
	}
	return account != nil && account.UserID == userID, nil
}
```

- [ ] **Step 4: Rodar e ver passar**

Run: `go test ./cmd/internal/services/transfers/ -v`
Expected: PASS (11 testes)

- [ ] **Step 5: Refatorar o handler**

Substituir `cmd/internal/handlers/transfers.go` por:

```go
// cmd/internal/handlers/transfers.go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	transferssvc "project_pay/cmd/internal/services/transfers"
)

type TransferHandler struct {
	svc *transferssvc.Service
}

func NewTransferHandler(svc *transferssvc.Service) *TransferHandler {
	return &TransferHandler{svc: svc}
}

func (h *TransferHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("body inválido"))
		return
	}

	transfer, err := h.svc.Create(r.Context(), req)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transfer)
}

func (h *TransferHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("id inválido"))
		return
	}

	transfer, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transfer)
}
```

- [ ] **Step 6: Wiring no main**

Em `cmd/api/main.go`, adicionar o import:

```go
	transferssvc "project_pay/cmd/internal/services/transfers"
```

e trocar:

```go
	transferRepo := transfersrepo.New(db)
	transferHandler := handlers.NewTransferHandler(transferRepo)
```

por:

```go
	transferRepo := transfersrepo.New(db)
	transferSvc := transferssvc.New(transferRepo, accountRepo)
	transferHandler := handlers.NewTransferHandler(transferSvc)
```

- [ ] **Step 7: Build + testes completos**

Run: `go build ./... ; go test ./...`
Expected: tudo verde.

- [ ] **Step 8: Commit**

```bash
git add cmd/internal/services/transfers/ cmd/internal/handlers/transfers.go cmd/api/main.go
git commit -m "feat: service de transfers com validacoes e ownership"
```

---

### Task 9: Services de participants e pix_keys (pass-through)

Sem regra de ownership no spec — o service só padroniza erros (`apperrors`) e mantém a arquitetura homogênea.

**Files:**
- Create: `cmd/internal/services/participants/participants.go`
- Test: `cmd/internal/services/participants/participants_test.go`
- Create: `cmd/internal/services/pix_keys/pix_keys.go`
- Test: `cmd/internal/services/pix_keys/pix_keys_test.go`
- Modify: `cmd/internal/handlers/participants.go`
- Modify: `cmd/internal/handlers/pix_keys.go`
- Modify: `cmd/api/main.go` (wiring)

- [ ] **Step 1: Testes de participants (falham)**

```go
// cmd/internal/services/participants/participants_test.go
package participants_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	participantssvc "project_pay/cmd/internal/services/participants"
)

type fakeRepo struct {
	participants map[uuid.UUID]*models.Participant
}

func (f *fakeRepo) Create(ctx context.Context, req models.CreateParticipantRequest) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Participant, error) {
	return f.participants[id], nil
}

func assertAppError(t *testing.T, err error, wantCode string) {
	t.Helper()
	var appErr *apperrors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("esperava *apperrors.Error, veio: %v", err)
	}
	if appErr.Code != wantCode {
		t.Errorf("code = %q, esperava %q", appErr.Code, wantCode)
	}
}

func TestGetByID_Inexistente404(t *testing.T) {
	svc := participantssvc.New(&fakeRepo{participants: map[uuid.UUID]*models.Participant{}})

	_, err := svc.GetByID(context.Background(), uuid.New())
	assertAppError(t, err, apperrors.CodeNotFound)
}

func TestGetByID_OK(t *testing.T) {
	id := uuid.New()
	svc := participantssvc.New(&fakeRepo{participants: map[uuid.UUID]*models.Participant{
		id: {ID: id, Name: "Banco X"},
	}})

	p, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if p.ID != id {
		t.Errorf("participant errado: %v", p.ID)
	}
}
```

- [ ] **Step 2: Testes de pix_keys (falham)**

```go
// cmd/internal/services/pix_keys/pix_keys_test.go
package pixkeys_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	pixkeyssvc "project_pay/cmd/internal/services/pix_keys"
)

type fakeRepo struct {
	keys map[string]*models.PixKey
}

func (f *fakeRepo) Create(ctx context.Context, req models.CreatePixKeysRequest) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (f *fakeRepo) GetByKey(ctx context.Context, key string) (*models.PixKey, error) {
	return f.keys[key], nil
}

func assertAppError(t *testing.T, err error, wantCode string) {
	t.Helper()
	var appErr *apperrors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("esperava *apperrors.Error, veio: %v", err)
	}
	if appErr.Code != wantCode {
		t.Errorf("code = %q, esperava %q", appErr.Code, wantCode)
	}
}

func TestGetByKey_Inexistente404(t *testing.T) {
	svc := pixkeyssvc.New(&fakeRepo{keys: map[string]*models.PixKey{}})

	_, err := svc.GetByKey(context.Background(), "chave@x.com")
	assertAppError(t, err, apperrors.CodeNotFound)
}

func TestGetByKey_OK(t *testing.T) {
	svc := pixkeyssvc.New(&fakeRepo{keys: map[string]*models.PixKey{
		"chave@x.com": {ID: uuid.New(), KeyValue: "chave@x.com"},
	}})

	k, err := svc.GetByKey(context.Background(), "chave@x.com")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if k.KeyValue != "chave@x.com" {
		t.Errorf("chave errada: %v", k.KeyValue)
	}
}
```

- [ ] **Step 3: Rodar e ver falhar**

Run: `go test ./cmd/internal/services/participants/ ./cmd/internal/services/pix_keys/ -v`
Expected: FAIL (pacotes não existem)

- [ ] **Step 4: Implementar os services**

```go
// cmd/internal/services/participants/participants.go
// Package participants contém a regra de negócio de participantes.
package participants

import (
	"context"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
)

type ParticipantRepository interface {
	Create(ctx context.Context, req models.CreateParticipantRequest) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Participant, error)
}

type Service struct {
	repo ParticipantRepository
}

func New(repo ParticipantRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req models.CreateParticipantRequest) (uuid.UUID, error) {
	if req.Ispb == "" || req.Name == "" {
		return uuid.Nil, apperrors.InvalidInput("ispb e name são obrigatórios")
	}

	id, err := s.repo.Create(ctx, req)
	if err != nil {
		return uuid.Nil, apperrors.Internal(err)
	}
	return id, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Participant, error) {
	participant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if participant == nil {
		return nil, apperrors.NotFound("participante não encontrado")
	}
	return participant, nil
}
```

```go
// cmd/internal/services/pix_keys/pix_keys.go
// Package pixkeys contém a regra de negócio de chaves pix.
package pixkeys

import (
	"context"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
)

type PixKeyRepository interface {
	Create(ctx context.Context, req models.CreatePixKeysRequest) (uuid.UUID, error)
	GetByKey(ctx context.Context, key string) (*models.PixKey, error)
}

type Service struct {
	repo PixKeyRepository
}

func New(repo PixKeyRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req models.CreatePixKeysRequest) (uuid.UUID, error) {
	if req.KeyValue == "" || req.KeyType == "" {
		return uuid.Nil, apperrors.InvalidInput("keyType e keyValue são obrigatórios")
	}

	id, err := s.repo.Create(ctx, req)
	if err != nil {
		return uuid.Nil, apperrors.Internal(err)
	}
	return id, nil
}

func (s *Service) GetByKey(ctx context.Context, key string) (*models.PixKey, error) {
	pixKey, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if pixKey == nil {
		return nil, apperrors.NotFound("chave pix não encontrada")
	}
	return pixKey, nil
}
```

Nota: confira os nomes dos campos de `models.CreateParticipantRequest` (`Ispb`, `Name`) e `models.CreatePixKeysRequest` (`KeyType`, `KeyValue`, `AccountID`) em `cmd/internal/models/participants.go` e `cmd/internal/models/pix_keys.go` antes de compilar — a validação usa esses campos.

- [ ] **Step 5: Rodar e ver passar**

Run: `go test ./cmd/internal/services/participants/ ./cmd/internal/services/pix_keys/ -v`
Expected: PASS (4 testes)

- [ ] **Step 6: Refatorar os handlers**

Substituir `cmd/internal/handlers/participants.go` por:

```go
// cmd/internal/handlers/participants.go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	participantssvc "project_pay/cmd/internal/services/participants"
)

type ParticipantHandler struct {
	svc *participantssvc.Service
}

func NewParticipantHandler(svc *participantssvc.Service) *ParticipantHandler {
	return &ParticipantHandler{svc: svc}
}

func (h *ParticipantHandler) CreateParticipant(w http.ResponseWriter, r *http.Request) {
	var req models.CreateParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("body inválido"))
		return
	}

	id, err := h.svc.Create(r.Context(), req)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.CreateParticipantResponse{ID: id})
}

func (h *ParticipantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("id inválido"))
		return
	}

	participant, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(participant)
}
```

Substituir `cmd/internal/handlers/pix_keys.go` por:

```go
// cmd/internal/handlers/pix_keys.go
package handlers

import (
	"encoding/json"
	"net/http"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	pixkeyssvc "project_pay/cmd/internal/services/pix_keys"
)

type PixKeyHandler struct {
	svc *pixkeyssvc.Service
}

func NewPixKeyHandler(svc *pixkeyssvc.Service) *PixKeyHandler {
	return &PixKeyHandler{svc: svc}
}

func (h *PixKeyHandler) CreatePixKey(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePixKeysRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("body inválido"))
		return
	}

	id, err := h.svc.Create(r.Context(), req)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.CreatePixKeysResponse{ID: id})
}

func (h *PixKeyHandler) GetByKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	pixKey, err := h.svc.GetByKey(r.Context(), key)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pixKey)
}
```

- [ ] **Step 7: Wiring no main**

Em `cmd/api/main.go`, adicionar os imports:

```go
	participantssvc "project_pay/cmd/internal/services/participants"
	pixkeyssvc "project_pay/cmd/internal/services/pix_keys"
```

e trocar os dois blocos de wiring:

```go
	participantRepo := participantsrepo.New(db)
	participantSvc := participantssvc.New(participantRepo)
	participantHandler := handlers.NewParticipantHandler(participantSvc)

	pixKeyRepo := pixkeysrepo.New(db)
	pixKeySvc := pixkeyssvc.New(pixKeyRepo)
	pixKeyHandler := handlers.NewPixKeyHandler(pixKeySvc)
```

- [ ] **Step 8: Build + testes completos**

Run: `go build ./... ; go test ./...`
Expected: tudo verde.

- [ ] **Step 9: Commit**

```bash
git add cmd/internal/services/participants/ cmd/internal/services/pix_keys/ cmd/internal/handlers/participants.go cmd/internal/handlers/pix_keys.go cmd/api/main.go
git commit -m "feat: services de participants e pix_keys no contrato apperrors"
```

---

### Task 10: Limpeza final + verificação de ponta a ponta

**Files:**
- Delete: `cmd/internal/services/services.go`
- Delete: `cmd/internal/repositories/repositories.go`
- Delete: `cmd/internal/handlers/handlers.go`

- [ ] **Step 1: Remover placeholders vazios**

```bash
rm cmd/internal/services/services.go cmd/internal/repositories/repositories.go cmd/internal/handlers/handlers.go
```

(`services/services.go` some porque agora cada entidade tem seu subpacote; os outros dois só declaravam o package e não têm conteúdo.)

- [ ] **Step 2: Build, vet e suite completa**

Run: `go build ./... ; go vet ./... ; go test ./...`
Expected: tudo verde. Se `handlers.go` ou `repositories.go` tiverem ganhado conteúdo em task anterior, NÃO remova — ajuste este step.

- [ ] **Step 3: Smoke manual de ownership (precisa do Postgres de pé)**

```bash
docker-compose up -d postgres
go run ./cmd/api
```

Em outro terminal (ajuste os IDs conforme as respostas):

```bash
# cria dois users
curl -s -X POST localhost:8080/users -d '{"name":"A","email":"a@x.com","tax_id":"1","password":"123"}'
curl -s -X POST localhost:8080/users -d '{"name":"B","email":"b@x.com","tax_id":"2","password":"123"}'

# login dos dois
curl -s -X POST localhost:8080/auth/login -d '{"email":"a@x.com","password":"123"}'   # => TOKEN_A
curl -s -X POST localhost:8080/auth/login -d '{"email":"b@x.com","password":"123"}'   # => TOKEN_B

# A cria participante e conta (body sem user_id!)
curl -s -X POST localhost:8080/participants -H "Authorization: Bearer TOKEN_A" -d '{"ispb":"12345678","name":"Banco Dev"}'
curl -s -X POST localhost:8080/accounts -H "Authorization: Bearer TOKEN_A" -d '{"participantID":"<PARTICIPANT_ID>","accountNumber":"1","branch":"0001","accountType":"cacc"}'

# A vê a própria conta => 200; B tenta ver a conta de A => 403 com JSON {"code":"FORBIDDEN",...}
curl -i localhost:8080/accounts/<ACC_A> -H "Authorization: Bearer TOKEN_A"
curl -i localhost:8080/accounts/<ACC_A> -H "Authorization: Bearer TOKEN_B"

# conta inexistente => 404; GET /accounts lista só as contas do token
curl -i localhost:8080/accounts/00000000-0000-0000-0000-00000000dead -H "Authorization: Bearer TOKEN_A"
curl -i localhost:8080/accounts -H "Authorization: Bearer TOKEN_A"
```

Expected: statuses 403/404/200 como indicado; todo response de erro no formato `{"code","message","trace_id"}`; header `X-Trace-Id` presente; logs JSON no stdout do servidor.

- [ ] **Step 4: Commit final**

```bash
git add -A
git commit -m "chore: remove placeholders vazios das camadas"
```

---

## Critério de aceite da Semana 3

- [ ] Nenhum `http.Error` sobrando em `cmd/internal/` (checar: `grep -r "http.Error" cmd/` retorna vazio).
- [ ] Todos os erros da API no formato `{"code","message","trace_id"}`.
- [ ] Logs JSON com trace_id por request; header `X-Trace-Id` em todo response.
- [ ] Handlers não importam mais `repositories/*` (exceto `handlers/auth.go`, mantido por decisão do spec).
- [ ] Ownership: 403 para recurso de outro user, 404 para inexistente.
- [ ] `go build ./... && go vet ./... && go test ./...` verdes.
