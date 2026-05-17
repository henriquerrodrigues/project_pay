package transfers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"project_pay/cmd/internal/models"
)

var ErrInsufficientFunds = errors.New("saldo insuficiente")

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// generateEndToEndID gera um id endtoend
func generateEndToEndID(ispb string) string {
	now := time.Now().UTC()
	date := now.Format("20060102")
	hhmm := now.Format("1504")
	suffix := strings.ReplaceAll(uuid.New().String(), "-", "")[:11]
	return fmt.Sprintf("E%s%s%s%s", ispb, date, hhmm, strings.ToUpper(suffix))
}

func (r *Repository) Create(ctx context.Context, req models.CreateTransferRequest) (*models.Transfer, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// buscar saldo e ispb do remetente
	var senderBalance int64
	var ispb string
	err = tx.QueryRow(ctx, sqlGetSenderInfo, req.SenderAccountID).Scan(&senderBalance, &ispb)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("conta remetente não encontrada")
	}
	if err != nil {
		return nil, err
	}

	if senderBalance < req.Amount {
		return nil, ErrInsufficientFunds
	}

	endToEndID := generateEndToEndID(ispb)

	// debitar remetente
	var senderBalanceAfter int64
	err = tx.QueryRow(ctx, sqlDebitAccount, req.Amount, req.SenderAccountID).Scan(&senderBalanceAfter)
	if err != nil {
		return nil, err
	}

	//buscar saldo do recebedor
	var receiverBalance int64
	err = tx.QueryRow(ctx, sqlGetReceiverBalance, req.ReceiverAccountID).Scan(&receiverBalance)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("conta recebedor não encontrada")
	}
	if err != nil {
		return nil, err
	}

	//creditar recebedor
	var receiverBalanceAfter int64
	err = tx.QueryRow(ctx, sqlCreditAccount, req.Amount, req.ReceiverAccountID).Scan(&receiverBalanceAfter)

	if err != nil {
		return nil, err
	}

	transfer := &models.Transfer{
		EndToendID:        endToEndID,
		SenderAccountID:   req.SenderAccountID,
		ReceiverAccountID: req.ReceiverAccountID,
		Amount:            req.Amount,
		Description:       req.Description,
	}
	err = tx.QueryRow(ctx, sqlCreateTransfer,
		endToEndID,
		req.SenderAccountID,
		req.ReceiverAccountID,
		req.Amount,
		req.Description,
	).Scan(
		&transfer.ID,
		&transfer.Status,
		&transfer.CreatedAt,
		&transfer.CompletedAt)
	if err != nil {
		return nil, err
	}

	//ledger_entry de DEBIT para o remetente
	_, err = tx.Exec(ctx, sqlCreateLedgerEntry,
		transfer.ID,
		req.SenderAccountID,
		req.Amount,
		"debit",
		senderBalanceAfter)

	if err != nil {
		return nil, err
	}

	//ledger_entry de CREDIT para o recebedor
	_, err = tx.Exec(ctx, sqlCreateLedgerEntry,
		transfer.ID,
		req.ReceiverAccountID,
		req.Amount,
		"credit",
		receiverBalanceAfter)

	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return transfer, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transfer, error) {
	transfer := &models.Transfer{}

	err := r.db.QueryRow(ctx, sqlBuscarTransfer, id).Scan(
		&transfer.ID,
		&transfer.EndToendID,
		&transfer.SenderAccountID,
		&transfer.ReceiverAccountID,
		&transfer.Amount,
		&transfer.Status,
		&transfer.Description,
		&transfer.CreatedAt,
		&transfer.CompletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return transfer, err
}

const sqlGetSenderInfo = `
	SELECT a.balance, p.ispb
	FROM accounts a
	JOIN participants p ON p.id = a.participant_id
	WHERE a.id = $1
	FOR UPDATE
`

const sqlDebitAccount = `
	UPDATE accounts
	set balance = balance - $1
	WHERE id = $2
	RETURNING balance
`

const sqlCreditAccount = `
	UPDATE accounts
	set balance = balance + $1
	WHERE id = $2
	RETURNING balance
`

const sqlGetReceiverBalance = `
	SELECT balance
	FROM accounts 
	WHERE id = $1
`

const sqlCreateTransfer = `
	INSERT INTO transfers (end_to_end_id, sender_account_id, receiver_account_id, amount, status, description, completed_at)
	VALUES ($1, $2, $3, $4, 'completed', $5, now())
	RETURNING id, status, created_at, completed_at`

const sqlCreateLedgerEntry = `
	INSERT INTO ledger_entries (transfer_id, account_id, amount, entry_type, balance_after)
	VALUES ($1, $2, $3, $4, $5)
`

const sqlBuscarTransfer = `
	SELECT id, end_to_end_id, sender_account_id, receiver_account_id, amount, status, description, created_at, completed_at
	FROM transfers
	WHERE id = $1
`
