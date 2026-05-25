package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"project_pay/cmd/internal/models"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, req models.CreateUserRequest) (uuid.UUID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.UUID{}, err
	}

	var id uuid.UUID
	err = r.db.QueryRow(ctx, sqlCreateUser,
		req.Name,
		req.Email,
		string(hash),
		req.TaxID,
	).Scan(&id)
	return id, err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, sqlGetByID, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.TaxID,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return user, err
}

const sqlCreateUser = `
	INSERT INTO users (name, email, password_hash, tax_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

const sqlGetByID = `
	SELECT id, name, email, tax_id, created_at
	FROM users
	WHERE id = $1`
