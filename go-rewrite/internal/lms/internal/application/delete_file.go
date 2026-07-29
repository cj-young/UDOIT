package application

import (
	"context"
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
	"rewritetest/internal/shared/auth"
)

type DeleteFileUseCase struct {
	lmsProviderRegistry         domain.LMSProviderRegistry
	lmsProviderConfigRepository domain.LMSProviderConfigRepository
	lmsObjectMappingRepository  domain.LMSObjectMappingRepository
}

func NewDeleteFileUseCase(
	lmsProviderRegistry domain.LMSProviderRegistry,
	lmsProviderConfigRepository domain.LMSProviderConfigRepository,
	lmsObjectMappingRepository domain.LMSObjectMappingRepository,
) *DeleteFileUseCase {
	return &DeleteFileUseCase{
		lmsProviderRegistry:         lmsProviderRegistry,
		lmsProviderConfigRepository: lmsProviderConfigRepository,
		lmsObjectMappingRepository:  lmsObjectMappingRepository,
	}
}

func (uc *DeleteFileUseCase) Execute(ctx context.Context, principal auth.Principal, fileID int64) error {
	lmsProvider, tenantConfig, err := GetLMSProviderAndConfig(
		ctx,
		uc.lmsProviderRegistry,
		uc.lmsProviderConfigRepository,
		principal.TenantID,
	)
	if err != nil {
		return err
	}

	fileMapping, err := uc.lmsObjectMappingRepository.GetByTypeAndInternalID(ctx, domain.LMSObjectTypeFile, fileID)
	if err != nil {
		return err
	}
	if fileMapping == nil {
		return apperr.New(
			apperr.CodeNotFound, "file_not_found", "LMS file not found",
			apperr.WithOp("lms.application.delete_file.Execute"),
		)
	}

	err = lmsProvider.DeleteFile(ctx, principal, tenantConfig, *fileMapping)
	if err != nil {
		return err
	}
	return nil
}
