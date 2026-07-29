package application

import (
	"context"
	"errors"
	"testing"

	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
	sharedauth "rewritetest/internal/shared/auth"

	"github.com/stretchr/testify/require"
)

var _ domain.FullLMSProvider = (*fakeLMSProvider)(nil)
var _ domain.LMSProviderRegistry = (*fakeLMSProviderRegistry)(nil)
var _ domain.LMSProviderConfigRepository = (*fakeLMSProviderConfigRepository)(nil)
var _ domain.LMSObjectMappingRepository = (*fakeObjectMappingRepository)(nil)

type fakeLMSProviderRegistry struct {
	provider domain.FullLMSProvider
	err error
	key domain.LMSType
}

func (f *fakeLMSProviderRegistry) Get(_ context.Context, lmsKey domain.LMSType) (domain.FullLMSProvider, error) {
	f.key = lmsKey
	if f.err != nil {
		return nil, f.err
	}
	return f.provider, nil
}

func (f *fakeLMSProviderRegistry) RegisterProvider(domain.LMSType, domain.FullLMSProvider) {}

type fakeLMSProviderConfigRepository struct {
	config *domain.LMSProviderConfig
	err    error
	called int
}

func (f *fakeLMSProviderConfigRepository) GetByTenant(context.Context, int64) (*domain.LMSProviderConfig, error) {
	f.called++
	if f.err != nil {
		return nil, f.err
	}
	return f.config, nil
}

func (f *fakeLMSProviderConfigRepository) UpsertByTenant(context.Context, int64, domain.LMSType, map[string]any) error {
	return nil
}

type fakeObjectMappingRepository struct {
	mapping *domain.LMSObjectMapping
	err     error
	called  int
}

func (f *fakeObjectMappingRepository) GetByTypeAndInternalID(_ context.Context, objectType domain.LMSObjectType, internalID int64) (*domain.LMSObjectMapping, error) {
	f.called++
	if f.err != nil {
		return nil, f.err
	}
	if f.mapping == nil {
		return nil, nil
	}
	if objectType != domain.LMSObjectTypeFile || internalID != 10 {
		return nil, errors.New("unexpected mapping lookup")
	}
	return f.mapping, nil
}

type fakeLMSProvider struct {
	err     error
	called  bool
	config  domain.LMSProviderConfig
	mapping domain.LMSObjectMapping
}

func (f *fakeLMSProvider) DeleteFile(_ context.Context, _ sharedauth.Principal, config domain.LMSProviderConfig, fileMapping domain.LMSObjectMapping) error {
	f.called = true
	f.config = config
	f.mapping = fileMapping
	return f.err
}

func (f *fakeLMSProvider) ValidateConfig(configData map[string]any) error {
	return nil
}

func (f *fakeLMSProvider) BeginAuthentication(ctx context.Context, providerConfig domain.LMSProviderConfig, userID int64, targetLinkURI string) (domain.AuthChallenge, error) {
	return domain.AuthChallenge{}, nil
}

func TestDeleteFileUseCase_MissingConfig(t *testing.T) {
	uc := NewDeleteFileUseCase(&fakeLMSProviderRegistry{}, &fakeLMSProviderConfigRepository{}, &fakeObjectMappingRepository{})

	err := uc.Execute(context.Background(), sharedauth.Principal{TenantID: 1}, 10)

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeInternal))
}

func TestDeleteFileUseCase_MissingProvider(t *testing.T) {
	config := domain.NewLMSProviderConfig(1, domain.LMSTypeCanvas, map[string]any{"baseUrl": "https://canvas.example"})
	uc := NewDeleteFileUseCase(&fakeLMSProviderRegistry{provider: nil}, &fakeLMSProviderConfigRepository{config: config}, &fakeObjectMappingRepository{})

	err := uc.Execute(context.Background(), sharedauth.Principal{TenantID: 1}, 10)

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeInternal))
}

func TestDeleteFileUseCase_MissingMapping(t *testing.T) {
	config := domain.NewLMSProviderConfig(1, domain.LMSTypeCanvas, map[string]any{"baseUrl": "https://canvas.example"})
	provider := &fakeLMSProvider{}
	uc := NewDeleteFileUseCase(
		&fakeLMSProviderRegistry{provider: provider},
		&fakeLMSProviderConfigRepository{config: config},
		&fakeObjectMappingRepository{},
	)

	err := uc.Execute(context.Background(), sharedauth.Principal{TenantID: 1}, 10)

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeNotFound))
	require.False(t, provider.called)
}

func TestDeleteFileUseCase_ProviderErrorBubbles(t *testing.T) {
	config := domain.NewLMSProviderConfig(1, domain.LMSTypeCanvas, map[string]any{"baseUrl": "https://canvas.example"})
	mapping := domain.NewLMSObjectMapping(10, domain.LMSObjectTypeFile, "canvas", "abc", map[string]any{"fileId": "abc"})
	provider := &fakeLMSProvider{err: errors.New("provider down")}
	uc := NewDeleteFileUseCase(
		&fakeLMSProviderRegistry{provider: provider},
		&fakeLMSProviderConfigRepository{config: config},
		&fakeObjectMappingRepository{mapping: &mapping},
	)

	err := uc.Execute(context.Background(), sharedauth.Principal{TenantID: 1}, 10)

	require.Error(t, err)
	require.EqualError(t, err, "provider down")
	require.True(t, provider.called)
}

func TestDeleteFileUseCase_Success(t *testing.T) {
	config := domain.NewLMSProviderConfig(1, domain.LMSTypeCanvas, map[string]any{"baseUrl": "https://canvas.example"})
	mapping := domain.NewLMSObjectMapping(10, domain.LMSObjectTypeFile, "canvas", "abc", map[string]any{"fileId": "abc"})
	provider := &fakeLMSProvider{}
	uc := NewDeleteFileUseCase(
		&fakeLMSProviderRegistry{provider: provider},
		&fakeLMSProviderConfigRepository{config: config},
		&fakeObjectMappingRepository{mapping: &mapping},
	)

	err := uc.Execute(context.Background(), sharedauth.Principal{TenantID: 1}, 10)

	require.NoError(t, err)
	require.True(t, provider.called)
	require.Equal(t, int64(10), provider.mapping.InternalID())
}
