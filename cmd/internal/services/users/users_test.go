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
