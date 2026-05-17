package transfers

import (
	"context"
	"errors"
	"project_pay/cmd/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, req models.CreateTransferRequest) (uuid.UUID, error) {
	var id uuid.UUID

	err := r.db.QueryRow(ctx, sqlCreateTransfer,
		req.EndToendID,
		req.SenderAccountID,
		req.ReceiverAccountID,
		req.Amount,
		req.Status)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transfer, error) {
	transfer := &models.Transfer{}

	err := r.db.QueryRow(ctx)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return transfer, err
}

const sqlCreateTransfer = `
	INSERT INTO transfers(end_to_end_id, sender_account_id, receiver_account_id, amount, status, description, created_at, completed_at)
	VALUES ($1, $2, $3, $4, $5, $6, now(), now())
`
