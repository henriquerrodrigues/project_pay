package accounts

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

func (r *Repository) Create(ctx context.Context, req models.CreateAccountRequest) (uuid.UUID, error) {
	var id uuid.UUID

	err := r.db.QueryRow(ctx, sqlCreateAccount,
		req.UserID,
		req.ParticipantID,
		req.AccountNumber,
		req.Branch,
		req.AccountType,
	).Scan(&id)

	return id, err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	account := &models.Account{}

	err := r.db.QueryRow(ctx, sqlGetByID, id).Scan(
		&account.ID,
		&account.UserID,
		&account.ParticipantID,
		&account.AccountNumber,
		&account.Branch,
		&account.AccountType,
		&account.Balance,
		&account.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return account, err
}

const sqlCreateAccount = `
	INSERT INTO accounts (user_id, participant_id, account_number, branch, account_type)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`

const sqlGetByID = `
	SELECT id, user_id, participant_id, account_number, branch, account_type, balance, created_at
	FROM accounts
	WHERE id = $1`
