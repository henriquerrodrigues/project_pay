// Package users contém a regra de negócio de usuários.
package users

import (
	"context"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/middleware"
	"project_pay/cmd/internal/models"
)

// UserRepository é o recorte do repository que o service precisa.
// A interface vive aqui para permitir fake nos testes.
type UserRepository interface {
	Create(ctx context.Context, req models.CreateUserRequest) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type Service struct {
	repo UserRepository
}

func New(repo UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req models.CreateUserRequest) (uuid.UUID, error) {
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return uuid.Nil, apperrors.InvalidInput("name, email e password são obrigatórios")
	}

	id, err := s.repo.Create(ctx, req)
	if err != nil {
		return uuid.Nil, apperrors.Internal(err)
	}
	return id, nil
}

// GetByID só permite o próprio usuário: recurso inexistente é 404,
// existente mas de outro user é 403.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	authUserID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.Unauthorized("usuário não autenticado")
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if user == nil {
		return nil, apperrors.NotFound("usuário não encontrado")
	}
	if user.ID != authUserID {
		return nil, apperrors.Forbidden("acesso negado")
	}
	return user, nil
}
