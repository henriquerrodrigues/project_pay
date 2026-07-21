// Package pixkeys contém a regra de negócio de chaves pix.
package pixkeys

import (
	"context"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
)

type PixKeyRepository interface {
	Create(ctx context.Context, req models.CreatePixKeysRequest) (uuid.UUID, error)
	GetByKey(ctx context.Context, key string) (*models.PixKey, error)
}

type Service struct {
	repo PixKeyRepository
}

func New(repo PixKeyRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req models.CreatePixKeysRequest) (uuid.UUID, error) {
	if req.KeyValue == "" || req.KeyType == "" {
		return uuid.Nil, apperrors.InvalidInput("keyType e keyValue são obrigatórios")
	}

	id, err := s.repo.Create(ctx, req)
	if err != nil {
		return uuid.Nil, apperrors.Internal(err)
	}
	return id, nil
}

func (s *Service) GetByKey(ctx context.Context, key string) (*models.PixKey, error) {
	pixKey, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if pixKey == nil {
		return nil, apperrors.NotFound("chave pix não encontrada")
	}
	return pixKey, nil
}
