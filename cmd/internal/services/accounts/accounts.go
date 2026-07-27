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

// Create cria uma conta vinculada ao user do token — o body não carrega user_id.
func (s *Service) Create(ctx context.Context, req models.CreateAccountRequest) (uuid.UUID, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return uuid.Nil, apperrors.Unauthorized("usuário não autenticado")
	}

	if req.ParticipantID == uuid.Nil || req.AccountNumber == "" || req.Branch == "" {
		return uuid.Nil, apperrors.InvalidInput("participantID, accountNumber e branch são obrigatórios")
	}
	if req.AccountType != "cacc" && req.AccountType != "svgs" {
		return uuid.Nil, apperrors.InvalidInput("accountType deve ser 'cacc' ou 'svgs'")
	}

	id, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return uuid.Nil, apperrors.Internal(err)
	}
	return id, nil
}

// GetByID aplica ownership: inexistente é 404, de outro user é 403.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if account == nil {
		return nil, apperrors.NotFound("conta não encontrada")
	}
	if account.UserID != userID {
		return nil, apperrors.Forbidden("conta não pertence ao usuário")
	}
	return account, nil
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
