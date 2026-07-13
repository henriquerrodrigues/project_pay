package participants_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	participantssvc "project_pay/cmd/internal/services/participants"
)

type fakeRepo struct {
	participants map[uuid.UUID]*models.Participant
}

func (f *fakeRepo) Create(ctx context.Context, req models.CreateParticipantRequest) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Participant, error) {
	return f.participants[id], nil
}

func assertAppError(t *testing.T, err error, wantCode string) {
	t.Helper()
	var appErr *apperrors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("esperava *apperrors.Error, veio: %v", err)
	}
	if appErr.Code != wantCode {
		t.Errorf("code = %q, esperava %q", appErr.Code, wantCode)
	}
}

func TestGetByID_Inexistente404(t *testing.T) {
	svc := participantssvc.New(&fakeRepo{participants: map[uuid.UUID]*models.Participant{}})

	_, err := svc.GetByID(context.Background(), uuid.New())
	assertAppError(t, err, apperrors.CodeNotFound)
}

func TestGetByID_OK(t *testing.T) {
	id := uuid.New()
	svc := participantssvc.New(&fakeRepo{participants: map[uuid.UUID]*models.Participant{
		id: {ID: id, Name: "Banco X"},
	}})

	p, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if p.ID != id {
		t.Errorf("participant errado: %v", p.ID)
	}
}
