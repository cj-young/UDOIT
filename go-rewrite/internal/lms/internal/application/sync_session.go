package application

import (
	"context"
	"strings"
	"time"

	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
)

type SyncSessionUseCase struct {
	mappingRepository    domain.UserLMSMappingRepository
	credentialRepository domain.LMSCredentialRepository
}

type SyncSessionCommand struct {
	UserID            int64
	LMSKey            string
	ExternalUserID    string
	APIDomain         string
	Metadata          map[string]any
	CredentialSchema  string
	CredentialPayload map[string]any
	CredentialExpires *time.Time
}

func NewSyncSessionUseCase(mappingRepository domain.UserLMSMappingRepository, credentialRepository domain.LMSCredentialRepository) *SyncSessionUseCase {
	return &SyncSessionUseCase{
		mappingRepository:    mappingRepository,
		credentialRepository: credentialRepository,
	}
}

func (u *SyncSessionUseCase) Execute(ctx context.Context, cmd SyncSessionCommand) error {
	if cmd.UserID <= 0 {
		return apperr.New(
			apperr.CodeValidation,
			"invalid_user_id",
			"user_id must be a positive integer",
			apperr.WithOp("lms.application.SyncSessionUseCase.Execute"),
		)
	}

	if strings.TrimSpace(cmd.LMSKey) == "" {
		return apperr.New(
			apperr.CodeValidation,
			"invalid_lms_key",
			"lms_key is required",
			apperr.WithOp("lms.application.SyncSessionUseCase.Execute"),
		)
	}

	now := time.Now()
	mapping := domain.NewUserLMSMapping(
		cmd.UserID,
		strings.TrimSpace(cmd.LMSKey),
		strings.TrimSpace(cmd.ExternalUserID),
		strings.TrimSpace(cmd.APIDomain),
		cmd.Metadata,
		now,
	)

	if err := u.mappingRepository.Upsert(ctx, mapping); err != nil {
		return err
	}

	if strings.TrimSpace(cmd.CredentialSchema) == "" || len(cmd.CredentialPayload) == 0 {
		return nil
	}

	credential := domain.NewLMSCredential(
		cmd.UserID,
		strings.TrimSpace(cmd.LMSKey),
		cmd.CredentialPayload,
		cmd.CredentialExpires,
		now,
	)

	return u.credentialRepository.UpsertActive(ctx, credential)
}
