// Package ledgerentries contém a regra de negócio do extrato.
package ledgerentries

import (
	"context"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"
)

type LedgerRepository interface {
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]models.LedgerEntry, error)
}

// AccountGetter é o recorte do repository de accounts usado pra checar ownership.
type AccountGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
}

type Service struct {
	repo     LedgerRepository
	accounts AccountGetter
}

func New(repo LedgerRepository, accounts AccountGetter) *Service {
	return &Service{repo: repo, accounts: accounts}
}

// GetStatement retorna o extrato; a conta deve pertencer ao user do token.
func (s *Service) GetStatement(ctx context.Context, accountID uuid.UUID) ([]models.LedgerEntry, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if account == nil {
		return nil, apperrors.NotFound("conta não encontrada")
	}
	if account.UserID != userID {
		return nil, apperrors.Forbidden("conta não pertence ao usuário")
	}

	entries, err := s.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return entries, nil
}
