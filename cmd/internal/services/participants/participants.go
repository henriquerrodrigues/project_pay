package participants

import (
	"context"
	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"

	"github.com/google/uuid"
)

type ParticipantRepository interface {
	Create(ctx context.Context, req models.CreateParticipantRequest) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Participant, error)
}

type Service struct {
	repo ParticipantRepository
}

func New(repo ParticipantRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req models.CreateParticipantRequest) (uuid.UUID, error) {
	if req.Ispb == "" || req.Name == "" {
		return uuid.Nil, apperrors.InvalidInput("ispb e name são obrigatórios")
	}

	id, err := s.repo.Create(ctx, req)
	if err != nil {
		return uuid.Nil, apperrors.Internal(err)
	}
	return id, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Participant, error) {
	participant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if participant == nil {
		return nil, apperrors.NotFound("participante não encontrado")
	}
	return participant, nil
}
