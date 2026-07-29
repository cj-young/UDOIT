package application

import (
	"context"
	"rewritetest/internal/lms/internal/domain"
)

type BeginAuthenticationUseCase struct {
	lmsProviderRegistry domain.LMSProviderRegistry
	lmsProviderConfigRepository domain.LMSProviderConfigRepository
}

func NewBeginAuthenticationUseCase(
	lmsProviderRegistry domain.LMSProviderRegistry,
	lmsProviderConfigRepository domain.LMSProviderConfigRepository,
) *BeginAuthenticationUseCase {
	return &BeginAuthenticationUseCase{
		lmsProviderRegistry: lmsProviderRegistry,
		lmsProviderConfigRepository: lmsProviderConfigRepository,
	}
}

func (u *BeginAuthenticationUseCase) Execute(ctx context.Context, userID int64, tenantID int64, targetLinkURI string) (domain.AuthChallenge, error) {

	providerConfig, err := u.lmsProviderConfigRepository.GetByTenant(ctx, tenantID)
	if err != nil {
		return domain.AuthChallenge{}, err
	}

	provider, err := u.lmsProviderRegistry.Get(ctx, providerConfig.LMSKey())
	if err != nil {
		return domain.AuthChallenge{}, err
	}

	return provider.BeginAuthentication(ctx, *providerConfig, userID, targetLinkURI)
}