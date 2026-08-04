package transfers

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"
	transfersrepo "project_pay/cmd/internal/repositories/transfers"
)

type TransferRepository interface {
	Create(ctx context.Context, req models.CreateTransferRequest) (*models.Transfer, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transfer, error)
}

type AccountGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
}

type Service struct {
	repo     TransferRepository
	accounts AccountGetter
}

func New(repo TransferRepository, accounts AccountGetter) *Service {
	return &Service{repo: repo, accounts: accounts}
}

func (s *Service) Create(ctx context.Context, req models.CreateTransferRequest) (*models.Transfer, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	if req.Amount <= 0 {
		return nil, apperrors.InvalidInput("valor deve ser maior que zero")
	}
	if req.SenderAccountID == req.ReceiverAccountID {
		return nil, apperrors.InvalidInput("conta de origem e destino não podem ser iguais")
	}

	sender, err := s.accounts.GetByID(ctx, req.SenderAccountID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if sender == nil {
		return nil, apperrors.NotFound("conta de origem não encontrada")
	}
	if sender.UserID != userID {
		return nil, apperrors.Forbidden("conta de origem não pertence ao usuário")
	}

	receiver, err := s.accounts.GetByID(ctx, req.ReceiverAccountID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if receiver == nil {
		return nil, apperrors.NotFound("conta de destino não encontrada")
	}

	transfer, err := s.repo.Create(ctx, req)
	if errors.Is(err, transfersrepo.ErrInsufficientFunds) {
		return nil, apperrors.InsufficientFunds()
	}
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return transfer, nil
}

// GetByID permite ver a transfer a quem for dono da conta origem OU destino.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Transfer, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	transfer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if transfer == nil {
		return nil, apperrors.NotFound("transferência não encontrada")
	}

	owns, err := s.userOwnsAccount(ctx, userID, transfer.SenderAccountID)
	if err != nil {
		return nil, err
	}
	if !owns {
		owns, err = s.userOwnsAccount(ctx, userID, transfer.ReceiverAccountID)
		if err != nil {
			return nil, err
		}
	}
	if !owns {
		return nil, apperrors.Forbidden("transferência não pertence ao usuário")
	}
	return transfer, nil
}

func (s *Service) userOwnsAccount(ctx context.Context, userID, accountID uuid.UUID) (bool, error) {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return false, apperrors.Internal(err)
	}
	return account != nil && account.UserID == userID, nil
}
