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

func (r *Repository) Create(ctx context.Context, req models.CreateAccountRequest, userID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID

	err := r.db.QueryRow(ctx, sqlCreateAccount,
		userID,
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

func (r *Repository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	rows, err := r.db.Query(ctx, sqlListByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []models.Account{}
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.ParticipantID,
			&a.AccountNumber,
			&a.Branch,
			&a.AccountType,
			&a.Balance,
			&a.CreatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

const sqlCreateAccount = `
	INSERT INTO accounts (user_id, participant_id, account_number, branch, account_type)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`

const sqlGetByID = `
	SELECT id, user_id, participant_id, account_number, branch, account_type, balance, created_at
	FROM accounts
	WHERE id = $1`

const sqlListByUserID = `
	SELECT id, user_id, participant_id, account_number, branch, account_type, balance, created_at
	FROM accounts`
