package application

import (
	"context"
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
)

type ProcessOAuthRedirectUseCase struct {
	lmsProviderRegistry domain.LMSProviderRegistry
	lmsProviderConfigRepository domain.LMSProviderConfigRepository
	authAttemptRepository domain.AuthAttemptRepository
}

func NewProcessOAuthRedirectUseCase(
	lmsProviderRegistry domain.LMSProviderRegistry,
	lmsProviderConfigRepository domain.LMSProviderConfigRepository,
	authAttemptRepository domain.AuthAttemptRepository,
) *ProcessOAuthRedirectUseCase {
	return &ProcessOAuthRedirectUseCase{
		lmsProviderRegistry: lmsProviderRegistry,
		lmsProviderConfigRepository: lmsProviderConfigRepository,
		authAttemptRepository: authAttemptRepository,
	}
}

type ProcessOAuthRedirectResponse struct {
	RedirectURL string
}

func (u *ProcessOAuthRedirectUseCase) Execute(ctx context.Context, state string, code string) (string, error) {
	
	authAttempt, err := u.authAttemptRepository.GetByState(ctx, state)
	if err != nil {
		return "", apperr.New(
			apperr.CodeInternal, "auth_attempt_not_found", "Failed to find auth attempt by state",
			apperr.WithOp("lms.internal.handler.handleOauthRedirect"),
			apperr.WithCause(err),
		)
	}

	providerConfig, err := u.lmsProviderConfigRepository.GetByTenant(ctx, authAttempt.TenantID)
	if err != nil {
		return "", apperr.New(
			apperr.CodeInternal, "provider_config_not_found", "Failed to find provider config by tenant ID",
			apperr.WithOp("lms.internal.handler.handleOauthRedirect"),
			apperr.WithCause(err),
		)
	}
	if providerConfig == nil {
		return "", apperr.New(
			apperr.CodeInternal, "provider_config_not_found", "Provider config is nil",
			apperr.WithOp("lms.internal.handler.handleOauthRedirect"),
		)
	}

	provider, err := u.lmsProviderRegistry.Get(ctx, providerConfig.LMSKey())
	if err != nil {
		return "", apperr.New(
			apperr.CodeInternal, "provider_not_found", "Failed to find provider by tenant ID",
			apperr.WithOp("lms.internal.handler.handleOauthRedirect"),
			apperr.WithCause(err),
		)
	}
	if provider == nil {
		return "", apperr.New(
			apperr.CodeInternal, "provider_not_found", "Provider is nil",
			apperr.WithOp("lms.internal.handler.handleOauthRedirect"),
		)
	}
	
	oauthRedirectProcessor, ok := provider.(domain.OAuthRedirectProcessor)
	if !ok {
		return "", apperr.New(
			apperr.CodeInternal, "provider_not_oauth", "The provider does not support OAuth",
			apperr.WithOp("lms.internal.handler.handleOauthRedirect"),
		)
	}

	redirectURL, err := oauthRedirectProcessor.ProcessOAuthRedirect(ctx, *providerConfig, authAttempt, code)
	if err != nil {
		return "", apperr.New(
			apperr.CodeInternal, "process_oauth_redirect_failed", "Failed to process OAuth redirect",
			apperr.WithOp("lms.internal.handler.handleOauthRedirect"),
			apperr.WithCause(err),
		)
	}

	return redirectURL, nil
}