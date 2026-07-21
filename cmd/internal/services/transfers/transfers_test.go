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
