package application

import (
	"context"
	"crypto/rand"
	"net/url"
	"time"

	"rewritetest/internal/lti/internal/domain"
	"rewritetest/internal/shared/apperr"
)

type GetLaunchRedirectUseCase struct {
	registrationRepository domain.RegistrationRepository
	ltiSessionRepository   domain.LTISessionRepository
}

type GetLaunchRedirectQuery struct {
	Issuer        string
	ClientID      string
	LoginHint     string
	TargetLinkURI string
	RedirectURI   string
	LTISessionTTL time.Duration
}

type LaunchRedirectParams struct {
	ClientID     string `json:"client_id"`
	State        string `json:"state"`
	Nonce        string `json:"nonce"`
	RedirectURI  string `json:"redirect_uri"`
	Scope        string `json:"scope"`
	ResponseType string `json:"response_type"`
	ResponseMode string `json:"response_mode"`
	Prompt       string `json:"prompt"`
}

func NewGetLaunchRedirectUseCase(registrationRepository domain.RegistrationRepository, ltiSessionRepository domain.LTISessionRepository) *GetLaunchRedirectUseCase {
	return &GetLaunchRedirectUseCase{
		registrationRepository: registrationRepository,
		ltiSessionRepository:   ltiSessionRepository,
	}
}

func (u *GetLaunchRedirectUseCase) Execute(ctx context.Context, query GetLaunchRedirectQuery) (string, error) {
	registration, err := u.registrationRepository.GetByIssuerAndClientID(ctx, query.Issuer, query.ClientID)
	if err != nil {
		return "", err
	}
	if registration == nil {
		return "", apperr.New(
			apperr.CodeNotFound,
			"registration_not_found",
			"LTI registration not found for the given issuer and client ID",
			apperr.WithOp("lti.application.GetLaunchRedirectUseCase.Execute"),
		)
	}

	baseURL, err := url.Parse(registration.LoginAuthEndpoint)
	if err != nil {
		return "", apperr.New(
			apperr.CodeInternal,
			"invalid_login_auth_endpoint",
			"Invalid login/auth endpoint URL in registration",
			apperr.WithOp("lti.application.GetLaunchRedirectUseCase.Execute"),
		)
	}

	state, err := generateState()
	if err != nil {
		return "", apperr.New(
			apperr.CodeInternal,
			"state_generation_failed",
			"Failed to generate state parameter",
			apperr.WithOp("lti.application.GetLaunchRedirectUseCase.Execute"),
		)
	}

	nonce, err := generateNonce()
	if err != nil {
		return "", apperr.New(
			apperr.CodeInternal,
			"nonce_generation_failed",
			"Failed to generate nonce parameter",
			apperr.WithOp("lti.application.GetLaunchRedirectUseCase.Execute"),
		)
	}

	params := baseURL.Query()
	params.Set("client_id", url.QueryEscape(registration.ClientID))
	params.Set("state", url.QueryEscape(state))
	params.Set("nonce", url.QueryEscape(nonce))
	params.Set("redirect_uri", url.QueryEscape(query.RedirectURI))
	params.Set("scope", "openid")
	params.Set("response_type", "id_token")
	params.Set("response_mode", "form_post")
	params.Set("prompt", "none")
	baseURL.RawQuery = params.Encode()

	ltiSession := domain.NewLTISession(state, nonce, query.Issuer, query.ClientID, query.TargetLinkURI, time.Now(), time.Now().Add(query.LTISessionTTL))
	if err := u.ltiSessionRepository.Create(ctx, ltiSession); err != nil {
		return "", apperr.New(
			apperr.CodeInternal,
			"lti_session_creation_failed",
			"Failed to create LTI session",
			apperr.WithOp("lti.application.GetLaunchRedirectUseCase.Execute"),
			apperr.WithCause(err),
		)
	}

	return baseURL.String(), nil
}

func generateState() (string, error) {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func generateNonce() (string, error) {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
