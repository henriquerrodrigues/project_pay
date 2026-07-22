package accounts

import (
	"context"
	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"

	"github.com/google/uuid"
)

type AccountRepository interface {
	Create(ctx context.Context, userID uuid.UUID, req models.CreateAccountRequest) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Account, error)
}

type Service struct {
	repo AccountRepository
}

func New(repo AccountRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListByCurrentUser(ctx context.Context) ([]models.Account, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	accounts, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return accounts, nil
}
