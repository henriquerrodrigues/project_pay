package pixkeys

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

func (r *Repository) Create(ctx context.Context, req models.CreatePixKeysRequest) (uuid.UUID, error) {
	var id uuid.UUID

	err := r.db.QueryRow(ctx, sqlCreatePixKey,
		req.KeyType,
		req.KeyValue,
		req.AccountID,
	).Scan(&id)

	return id, err
}

func (r *Repository) GetByKey(ctx context.Context, key string) (*models.PixKey, error) {
	pix_key := &models.PixKey{}

	err := r.db.QueryRow(ctx, sqlBuscarPixKey, key).Scan(
		&pix_key.ID,
		&pix_key.KeyType,
		&pix_key.KeyValue,
		&pix_key.AccountID,
		&pix_key.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return pix_key, err
}

const sqlCreatePixKey = `
	INSERT INTO pix_keys (key_type, key_value, account_id)
	VALUES ($1, $2, $3)
	RETURNING id`

const sqlBuscarPixKey = `
	SELECT id, key_type, key_value, account_id, created_at
	FROM pix_keys
	WHERE key_value = $1`
