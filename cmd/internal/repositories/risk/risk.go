package risk

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"project_pay/cmd/internal/models"
	"project_pay/cmd/internal/risk"
)

var ErrAccountNotFound = errors.New("conta não encontrada")

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// LoadSignals monta os sinais que os agentes de risco analisam.
// Só lê o banco; não trava nenhuma linha.
func (r *Repository) LoadSignals(ctx context.Context, req models.CreateTransferRequest, now time.Time) (risk.Signals, error) {
	s := risk.Signals{Amount: req.Amount, Now: now}

	err := r.db.QueryRow(ctx, sqlSenderBalance, req.SenderAccountID).Scan(&s.SenderBalance)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrAccountNotFound
	}
	if err != nil {
		return s, err
	}

	var receiverCreatedAt time.Time
	err = r.db.QueryRow(ctx, sqlReceiverCreatedAt, req.ReceiverAccountID).Scan(&receiverCreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrAccountNotFound
	}
	if err != nil {
		return s, err
	}
	s.ReceiverAccountAge = now.Sub(receiverCreatedAt)

	err = r.db.QueryRow(ctx, sqlSenderHistory, req.SenderAccountID, now).Scan(
		&s.TransfersLast10Min,
		&s.HistoryCount,
		&s.AvgAmount,
	)
	if err != nil {
		return s, err
	}

	var previous int
	err = r.db.QueryRow(ctx, sqlPreviousToReceiver, req.SenderAccountID, req.ReceiverAccountID).Scan(&previous)
	if err != nil {
		return s, err
	}
	s.FirstTimeReceiver = previous == 0

	return s, nil
}

// Save grava a decisão na trilha de auditoria. transferID é nil quando a
// transferência não foi executada (bloqueada, em revisão ou com erro).
func (r *Repository) Save(ctx context.Context, req models.CreateTransferRequest, transferID *uuid.UUID, a risk.Assessment) error {
	_, err := r.db.Exec(ctx, sqlSaveDecision,
		transferID,
		req.SenderAccountID,
		req.ReceiverAccountID,
		req.Amount,
		string(a.Decision),
		a.FraudScore,
		string(a.FraudLevel),
		a.RedFlags,
		a.Compliance,
		a.Explanation,
	)
	return err
}

const sqlSenderBalance = `
	SELECT balance
	FROM accounts
	WHERE id = $1
`

const sqlReceiverCreatedAt = `
	SELECT created_at
	FROM accounts
	WHERE id = $1
`

const sqlSenderHistory = `
	SELECT
		count(*) FILTER (WHERE created_at >= $2::timestamptz - interval '10 minutes'),
		count(*),
		coalesce(avg(amount), 0)::bigint
	FROM transfers
	WHERE sender_account_id = $1
	  AND status = 'completed'
	  AND created_at >= $2::timestamptz - interval '30 days'
`

const sqlPreviousToReceiver = `
	SELECT count(*)
	FROM transfers
	WHERE sender_account_id = $1
	  AND receiver_account_id = $2
	  AND status = 'completed'
`

const sqlSaveDecision = `
	INSERT INTO risk_decisions (transfer_id, sender_account_id, receiver_account_id, amount,
		decision, fraud_score, fraud_level, red_flags, compliance_violations, explanation)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`
