package participants

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

func (r *Repository) Create(ctx context.Context, req models.CreateParticipantRequest) (uuid.UUID, error) {
	var id uuid.UUID

	err := r.db.QueryRow(ctx, sqlCreateParticipant,
		req.Ispb,
		req.Name,
	).Scan(&id)

	return id, err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Participant, error) {
	participant := &models.Participant{}

	err := r.db.QueryRow(ctx, sqlBuscarIDParticipant, id).Scan(
		&participant.ID,
		&participant.Ispb,
		&participant.Name,
		&participant.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return participant, err
}

const sqlCreateParticipant = `
	INSERT INTO participants (ispb, name)
	VALUES ($1, $2)
	RETURNING id`

const sqlBuscarIDParticipant = `
	SELECT id, ispb, name, created_at
	FROM participants
	WHERE id = $1`
