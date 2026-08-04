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
