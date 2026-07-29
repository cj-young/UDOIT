package application

import (
	"context"
	"rewritetest/internal/auth/internal/domain"
	"time"

	"github.com/google/uuid"
)

type CreateSessionUseCase struct {
	sessionRepository domain.SessionRepository
}

type CreateSessionCommand struct {
	UserID int64
	TenantID int64
	TTL time.Duration
}

func NewCreateSessionUseCase(sessionRepository domain.SessionRepository) *CreateSessionUseCase {
	return &CreateSessionUseCase{
		sessionRepository: sessionRepository,
	}
}

func (u *CreateSessionUseCase) Execute(ctx context.Context, cmd CreateSessionCommand) (string, error) {
	sessionID := uuid.NewString()
	session := domain.NewSession(sessionID, cmd.UserID, cmd.TenantID, time.Now(), time.Now().Add(cmd.TTL))
	if err := u.sessionRepository.Create(ctx, session); err != nil {
		return "", err
	}

	return session.ID(), nil
}