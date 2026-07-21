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
