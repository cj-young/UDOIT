package infrastructure

import (
	"context"
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
)

type MapLMSProviderRegistry struct {
	providers map[string]domain.FullLMSProvider
}

func NewMapLMSProviderRegistry() *MapLMSProviderRegistry {
	return &MapLMSProviderRegistry{
		providers: make(map[string]domain.FullLMSProvider),
	}
}

func (r *MapLMSProviderRegistry) RegisterProvider(lmsType domain.LMSType, provider domain.FullLMSProvider) {
	r.providers[string(lmsType)] = provider
}

func (r *MapLMSProviderRegistry) Get(ctx context.Context, lmsType domain.LMSType) (domain.FullLMSProvider, error) {
	provider, exists := r.providers[string(lmsType)]
	if !exists {
		return nil, apperr.New(
			apperr.CodeInternal, "invalid_provider", "An invalid provider key was supplied.",
		)
	}
	return provider, nil
}

var _ domain.LMSProviderRegistry = (*MapLMSProviderRegistry)(nil)