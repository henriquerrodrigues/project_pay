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
