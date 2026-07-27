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

func TestCreate_SemUserNoContext401(t *testing.T) {
	svc := accountssvc.New(&fakeRepo{})

	_, err := svc.Create(context.Background(), validReq())
	assertAppError(t, err, apperrors.CodeUnauthorized)
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

func TestGetByID_SemUserNoContext401(t *testing.T) {
	svc := accountssvc.New(&fakeRepo{accounts: map[uuid.UUID]*models.Account{}})

	_, err := svc.GetByID(context.Background(), uuid.New())
	assertAppError(t, err, apperrors.CodeUnauthorized)
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
