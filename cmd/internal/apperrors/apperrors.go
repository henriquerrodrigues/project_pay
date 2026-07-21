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
