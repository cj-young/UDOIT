package application

import (
	"context"

	"rewritetest/internal/lti/internal/domain"
	"rewritetest/internal/shared/apperr"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

type ProcessLaunchUseCase struct {
	ltiSessionRepository   domain.LTISessionRepository
	registrationRepository domain.RegistrationRepository
	ltiUserLinkRepository  domain.LTIUserLinkRepository
	userCreator            UserCreator
}

type UserCreator interface {
	CreateUser(ctx context.Context, username string, name string) (int64, error)
}

type ProcessLaunchCommand struct {
	IDToken string
	State   string
}

type ProcessLaunchResult struct {
	RedirectURL string
	UserID      int64
}

func NewProcessLaunchUseCase(
	ltiSessionRepository domain.LTISessionRepository,
	registrationRepository domain.RegistrationRepository,
	ltiUserLinkRepository domain.LTIUserLinkRepository,
	userCreator UserCreator,
) *ProcessLaunchUseCase {
	return &ProcessLaunchUseCase{
		ltiSessionRepository:   ltiSessionRepository,
		registrationRepository: registrationRepository,
		ltiUserLinkRepository:  ltiUserLinkRepository,
		userCreator:            userCreator,
	}
}

func (u *ProcessLaunchUseCase) Execute(ctx context.Context, cmd ProcessLaunchCommand) (ProcessLaunchResult, error) {
	session, err := u.ltiSessionRepository.GetByState(ctx, cmd.State)
	if err != nil {
		return ProcessLaunchResult{}, err
	}

	if session == nil || session.IsExpired() {
		return ProcessLaunchResult{}, apperr.New(
			apperr.CodeNotFound,
			"session_not_found",
			"LTI session not found or expired",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
		)
	}

	registration, err := u.registrationRepository.GetByIssuerAndClientID(ctx, session.Issuer(), session.ClientID())
	if err != nil {
		return ProcessLaunchResult{}, err
	}
	if registration == nil {
		return ProcessLaunchResult{}, apperr.New(
			apperr.CodeNotFound,
			"registration_not_found",
			"LTI registration not found for issuer and client ID",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
		)
	}

	unverified, _, err := jwt.NewParser().ParseUnverified(cmd.IDToken, jwt.MapClaims{})
	if err != nil {
		return ProcessLaunchResult{}, apperr.New(
			apperr.CodeInternal,
			"invalid_id_token",
			"Failed to parse ID token",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
			apperr.WithCause(err),
		)
	}

	kid, ok := unverified.Header["kid"].(string)
	if !ok || kid == "" {
		return ProcessLaunchResult{}, apperr.New(
			apperr.CodeInternal,
			"invalid_id_token",
			"ID token missing 'kid' header",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
		)
	}

	k, err := keyfunc.NewDefaultCtx(ctx, []string{registration.JWKEndpoint})
	if err != nil {
		return ProcessLaunchResult{}, apperr.New(
			apperr.CodeInternal,
			"jwks_client_error",
			"Failed to create JWKS client",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
			apperr.WithCause(err),
		)
	}

	parsed, err := jwt.Parse(cmd.IDToken, k.Keyfunc, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil || !parsed.Valid {
		return ProcessLaunchResult{}, apperr.New(
			apperr.CodeInternal,
			"invalid_id_token",
			"ID token validation failed",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
			apperr.WithCause(err),
		)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return ProcessLaunchResult{}, apperr.New(
			apperr.CodeInternal,
			"invalid_id_token",
			"Invalid ID token claims",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
		)
	}

	if err := validateLaunchClaims(claims, session); err != nil {
		return ProcessLaunchResult{}, err
	}

	if err := u.ltiSessionRepository.Delete(ctx, cmd.State); err != nil {
		return ProcessLaunchResult{}, err
	}

	link, err := u.ltiUserLinkRepository.GetBySubAndIssuer(ctx, claims["sub"].(string), session.Issuer())
	if err != nil {
		return ProcessLaunchResult{}, err
	}
	if link == nil {
		userID, err := u.userCreator.CreateUser(ctx, claims["sub"].(string), claims["name"].(string))
		if err != nil {
			return ProcessLaunchResult{}, err
		}
		link = domain.NewLTIUserLink(claims["sub"].(string), session.Issuer(), userID)
		if err := u.ltiUserLinkRepository.Create(ctx, link); err != nil {
			return ProcessLaunchResult{}, err
		}
	}

	return ProcessLaunchResult{
		RedirectURL: session.TargetLinkURI(),
		UserID:      link.UserID(),
	}, nil
}

func validateLaunchClaims(claims jwt.MapClaims, session *domain.LTISession) error {
	if iss, _ := claims["iss"].(string); iss != session.Issuer() {
		return apperr.New(
			apperr.CodeValidation,
			"invalid_id_token",
			"ID token 'iss' claim does not match session issuer",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.validateLaunchClaims"),
		)
	}

	if nonce, _ := claims["nonce"].(string); nonce != session.Nonce() {
		return apperr.New(
			apperr.CodeValidation,
			"invalid_id_token",
			"ID token 'nonce' claim does not match session nonce",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.validateLaunchClaims"),
		)
	}

	if !audienceContains(claims["aud"], session.ClientID()) {
		return apperr.New(
			apperr.CodeValidation,
			"invalid_id_token",
			"ID token 'aud' claim does not contain session client ID",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.validateLaunchClaims"),
		)
	}

	return nil
}

func audienceContains(aud any, clientID string) bool {
	switch v := aud.(type) {
	case string:
		return v == clientID
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s == clientID {
				return true
			}
		}
	}
	return false
}
