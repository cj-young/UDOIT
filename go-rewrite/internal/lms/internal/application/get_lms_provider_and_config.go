package application

import (
	"context"

	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
)

func GetLMSProviderAndConfig(
	ctx context.Context,
	lmsProviderRegistry domain.LMSProviderRegistry,
	lmsProviderConfigRepository domain.LMSProviderConfigRepository,
	tenantID int64,
) (domain.FullLMSProvider, domain.LMSProviderConfig, error) {
	tenantConfig, err := lmsProviderConfigRepository.GetByTenant(ctx, tenantID)

	var zero domain.FullLMSProvider

	if err != nil {
		return zero, domain.LMSProviderConfig{}, err
	}
	if tenantConfig == nil {
		return zero, domain.LMSProviderConfig{}, apperr.New(
			apperr.CodeInternal, "missing_lms_provider_config", "LMS provider config not found for tenant",
			apperr.WithOp("lms.application.get_lms_provider_and_config"),
		)
	}

	lmsProvider, err := lmsProviderRegistry.Get(ctx, tenantConfig.LMSKey())
	if err != nil {
		return zero, domain.LMSProviderConfig{}, err
	}
	
	return lmsProvider, *tenantConfig, nil
}