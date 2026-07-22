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

type fakeLMSProviderRegistry struct {
	provider domain.LMSProvider
	err      error
	key      domain.LMSType
}

func (f *fakeLMSProviderRegistry) Get(_ context.Context, lmsKey domain.LMSType) (domain.LMSProvider, error) {
	f.key = lmsKey
	if f.err != nil {
		return nil, f.err
	}
	return f.provider, nil
}

func (f *fakeLMSProviderRegistry) RegisterProvider(domain.LMSType, domain.LMSProvider) {}

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
	mapping := domain.NewLMSObjectMapping(10, domain.LMSObjectTypeFile, "canvas", map[string]any{"fileId": "abc"})
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
	mapping := domain.NewLMSObjectMapping(10, domain.LMSObjectTypeFile, "canvas", map[string]any{"fileId": "abc"})
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
