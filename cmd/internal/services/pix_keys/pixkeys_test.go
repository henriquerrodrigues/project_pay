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
