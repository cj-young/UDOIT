package lms

import (
	"context"
	"rewritetest/internal/lms/internal/application"
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
	"rewritetest/internal/shared/auth"
)

func (m *Module) DeleteFile(ctx context.Context, principal auth.Principal, fileID int64) error {
	lmsProvider, tenantConfig, err := application.GetLMSProviderAndConfig(
		ctx,
		m.providerRegistry,
		m.providerConfigRepository,
		principal.TenantID,
	)
	if err != nil {
		return err
	}

	fileMapping, err := m.objectMappingRepository.GetByTypeAndInternalID(ctx, domain.LMSObjectTypeFile, fileID)
	if err != nil {
		return err
	}
	if fileMapping == nil {
		return apperr.New(
			apperr.CodeNotFound, "file_not_found", "LMS file not found",
		)
	}

	err = lmsProvider.DeleteFile(ctx, principal, tenantConfig, *fileMapping)
	if err != nil {
		return err
	}
	return nil
}

