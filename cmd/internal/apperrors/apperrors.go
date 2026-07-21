package apperrors

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"project_pay/cmd/internal/traceid"
)

const (
	CodeNotFound          = "NOT_FOUND"
	CodeInvalidInput      = "INVALID_INPUT"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	CodeInternal          = "INTERNAL"
)

type Error struct {
	Code       string
	HTTPStatus int
	Message    string
	err        error //causa interna; não enviar no response
}

type response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

// nome reservador do contrato com o pacote errors
func (e *Error) Error() string {
	if e.err != nil {
		return e.Code + ": " + e.err.Error()
	}
	return e.Code + ": " + e.Message
}

// nome reservador do contrato com o pacote errors
func (e *Error) Unwrap() error { return e.err }

func NotFound(msg string) *Error {
	return &Error{
		Code:       CodeNotFound,
		HTTPStatus: http.StatusNotFound,
		Message:    msg,
	}
}

func InvalidInput(msg string) *Error {
	return &Error{
		Code:       CodeInvalidInput,
		HTTPStatus: http.StatusBadRequest,
		Message:    msg,
	}
}

func Unauthorized(msg string) *Error {
	return &Error{
		Code:       CodeUnauthorized,
		HTTPStatus: http.StatusUnauthorized,
		Message:    msg,
	}
}

func Forbidden(msg string) *Error {
	return &Error{
		Code:       CodeForbidden,
		HTTPStatus: http.StatusForbidden,
		Message:    msg,
	}
}

func InsufficientFunds() *Error {
	return &Error{
		Code:       CodeInsufficientFunds,
		HTTPStatus: http.StatusUnprocessableEntity,
		Message:    "saldo insuficiente",
	}
}

func Internal(err error) *Error {
	return &Error{
		Code:       CodeInternal,
		HTTPStatus: http.StatusInternalServerError,
		Message:    "erro interno",
		err:        err,
	}
}

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
