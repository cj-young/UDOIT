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
	tenantConfig, err := uc.lmsProviderConfigRepository.GetByTenant(ctx, principal.TenantID)
	if err != nil {
		return err
	}
	if tenantConfig == nil {
		return apperr.New(
			apperr.CodeInternal, "missing_lms_provider_config", "LMS provider config not found for tenant",
			apperr.WithOp("lms.application.delete_file.Execute"),
		)
	}

	lmsProvider, err := uc.lmsProviderRegistry.Get(ctx, tenantConfig.LMSKey())
	if err != nil {
		return err
	}
	if lmsProvider == nil {
		return apperr.New(
			apperr.CodeInternal, "missing_lms_provider", "LMS provider not found for tenant",
			apperr.WithOp("lms.application.delete_file.Execute"),
		)
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

	return lmsProvider.DeleteFile(ctx, principal, *tenantConfig, *fileMapping)
}
