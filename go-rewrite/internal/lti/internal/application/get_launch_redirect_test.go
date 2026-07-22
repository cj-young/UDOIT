package application

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"rewritetest/internal/lti/internal/domain"
	"rewritetest/internal/shared/apperr"

	"github.com/stretchr/testify/require"
)

type fakeRedirectRegistrationRepository struct {
	registration *domain.Registration
	err          error
	issuer       string
	clientID     string
}

func (f *fakeRedirectRegistrationRepository) Create(context.Context, domain.Registration) error {
	return nil
}

func (f *fakeRedirectRegistrationRepository) Save(context.Context, domain.Registration) error {
	return nil
}

func (f *fakeRedirectRegistrationRepository) GetByIssuerAndClientID(_ context.Context, issuer string, clientID string) (*domain.Registration, error) {
	f.issuer = issuer
	f.clientID = clientID
	if f.err != nil {
		return nil, f.err
	}
	return f.registration, nil
}

type fakeRedirectSessionRepository struct {
	createErr error
	created   []*domain.LTISession
}

func (f *fakeRedirectSessionRepository) Create(_ context.Context, session *domain.LTISession) error {
	f.created = append(f.created, session)
	return f.createErr
}

func (f *fakeRedirectSessionRepository) GetByState(context.Context, string) (*domain.LTISession, error) {
	return nil, nil
}

func (f *fakeRedirectSessionRepository) Delete(context.Context, string) error {
	return nil
}

func TestGetLaunchRedirectUseCase_Execute_RegistrationNotFound_ReturnsNotFound(t *testing.T) {
	registrationRepo := &fakeRedirectRegistrationRepository{}
	sessionRepo := &fakeRedirectSessionRepository{}
	uc := NewGetLaunchRedirectUseCase(registrationRepo, sessionRepo)

	_, err := uc.Execute(context.Background(), GetLaunchRedirectQuery{
		Issuer:        "https://issuer.example",
		ClientID:      "client-123",
		LoginHint:     "hint",
		TargetLinkURI: "https://tool.example/app",
		RedirectURI:   "https://tool.example/lti/authorize/check",
	})

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeNotFound))
	require.Empty(t, sessionRepo.created)
}

func TestGetLaunchRedirectUseCase_Execute_InvalidLoginAuthEndpoint_ReturnsInternal(t *testing.T) {
	registrationRepo := &fakeRedirectRegistrationRepository{
		registration: domain.NewRegistration(
			"https://issuer.example",
			"client-123",
			"://invalid-url",
			"https://issuer.example/jwks",
			"https://issuer.example/service-auth",
			"https://issuer.example/logout",
			2,
		),
	}
	sessionRepo := &fakeRedirectSessionRepository{}
	uc := NewGetLaunchRedirectUseCase(registrationRepo, sessionRepo)

	_, err := uc.Execute(context.Background(), GetLaunchRedirectQuery{
		Issuer:        "https://issuer.example",
		ClientID:      "client-123",
		LoginHint:     "hint",
		TargetLinkURI: "https://tool.example/app",
		RedirectURI:   "https://tool.example/lti/authorize/check",
	})

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeInternal))
	require.Empty(t, sessionRepo.created)
}

func TestGetLaunchRedirectUseCase_Execute_CreateSessionFails_ReturnsInternal(t *testing.T) {
	registrationRepo := &fakeRedirectRegistrationRepository{
		registration: domain.NewRegistration(
			"https://issuer.example",
			"client-123",
			"https://issuer.example/oidc/auth",
			"https://issuer.example/jwks",
			"https://issuer.example/service-auth",
			"https://issuer.example/logout",
			2,
		),
	}
	sessionRepo := &fakeRedirectSessionRepository{createErr: errors.New("redis unavailable")}
	uc := NewGetLaunchRedirectUseCase(registrationRepo, sessionRepo)

	_, err := uc.Execute(context.Background(), GetLaunchRedirectQuery{
		Issuer:        "https://issuer.example",
		ClientID:      "client-123",
		LoginHint:     "hint",
		TargetLinkURI: "https://tool.example/app",
		RedirectURI:   "https://tool.example/lti/authorize/check",
	})

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeInternal))
	require.Len(t, sessionRepo.created, 1)
}

func TestGetLaunchRedirectUseCase_Execute_UsesDefaultTTLWhenMissing(t *testing.T) {
	registrationRepo := &fakeRedirectRegistrationRepository{
		registration: domain.NewRegistration(
			"https://issuer.example",
			"client-123",
			"https://issuer.example/oidc/auth",
			"https://issuer.example/jwks",
			"https://issuer.example/service-auth",
			"https://issuer.example/logout",
			2,
		),
	}
	sessionRepo := &fakeRedirectSessionRepository{}
	uc := NewGetLaunchRedirectUseCase(registrationRepo, sessionRepo)

	_, err := uc.Execute(context.Background(), GetLaunchRedirectQuery{
		Issuer:        "https://issuer.example",
		ClientID:      "client-123",
		LoginHint:     "hint",
		TargetLinkURI: "https://tool.example/app",
		RedirectURI:   "https://tool.example/lti/authorize/check",
	})

	require.NoError(t, err)
	require.Len(t, sessionRepo.created, 1)
	require.False(t, sessionRepo.created[0].IsExpired())
}

func TestGetLaunchRedirectUseCase_Execute_Success_PersistsSessionAndBuildsOIDCRedirect(t *testing.T) {
	registration := domain.NewRegistration(
		"https://issuer.example",
		"client-123",
		"https://issuer.example/oidc/auth",
		"https://issuer.example/jwks",
		"https://issuer.example/service-auth",
		"https://issuer.example/logout",
		2,
	)
	registrationRepo := &fakeRedirectRegistrationRepository{registration: registration}
	sessionRepo := &fakeRedirectSessionRepository{}
	uc := NewGetLaunchRedirectUseCase(registrationRepo, sessionRepo)

	redirectURI := "https://tool.example/lti/authorize/check"
	result, err := uc.Execute(context.Background(), GetLaunchRedirectQuery{
		Issuer:        "https://issuer.example",
		ClientID:      "client-123",
		LoginHint:     "hint",
		TargetLinkURI: "https://tool.example/app",
		RedirectURI:   redirectURI,
		LTISessionTTL: 5 * time.Minute,
	})

	require.NoError(t, err)
	require.Equal(t, "https://issuer.example", registrationRepo.issuer)
	require.Equal(t, "client-123", registrationRepo.clientID)
	require.Len(t, sessionRepo.created, 1)

	parsed, parseErr := url.Parse(result)
	require.NoError(t, parseErr)
	require.Equal(t, "https", parsed.Scheme)
	require.Equal(t, "issuer.example", parsed.Host)
	require.Equal(t, "/oidc/auth", parsed.Path)

	params := parsed.Query()
	require.Equal(t, "client-123", params.Get("client_id"))
	require.Equal(t, redirectURI, params.Get("redirect_uri"))
	require.Equal(t, "openid", params.Get("scope"))
	require.Equal(t, "id_token", params.Get("response_type"))
	require.Equal(t, "form_post", params.Get("response_mode"))
	require.Equal(t, "none", params.Get("prompt"))
	require.NotEmpty(t, params.Get("state"))
	require.NotEmpty(t, params.Get("nonce"))

	createdSession := sessionRepo.created[0]
	require.Equal(t, "https://issuer.example", createdSession.Issuer())
	require.Equal(t, "client-123", createdSession.ClientID())
	require.Equal(t, int64(2), createdSession.TenantID())
	require.Equal(t, "https://tool.example/app", createdSession.TargetLinkURI())
	require.Equal(t, params.Get("state"), createdSession.State())
	require.Equal(t, params.Get("nonce"), createdSession.Nonce())
	require.False(t, createdSession.IsExpired())
}
