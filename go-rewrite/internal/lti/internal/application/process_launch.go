package application

import (
	"context"

	"rewritetest/internal/lti/internal/domain"
	"rewritetest/internal/shared/apperr"
	"rewritetest/internal/tenants"

	"github.com/golang-jwt/jwt/v5"
)

type ProcessLaunchUseCase struct {
	ltiSessionRepository    	domain.LTISessionRepository
	registrationRepository  	domain.RegistrationRepository
	ltiUserLinkRepository   	domain.LTIUserLinkRepository
	ltiCourseLinkRepository 	domain.LTICourseLinkRepository

	tenantRetriever           TenantRetriever
	idTokenVerifier         	IDTokenVerifier
	userCreator             	UserCreator
	courseCreator           	CourseCreator
	courseInfoRetriever     	CourseInfoRetriever
}

type UserCreator interface {
	CreateUser(ctx context.Context, username string, name string) (int64, error)
}

type CourseCreator interface {
	CreateCourse(ctx context.Context, title string, tenantID int64, externalID string, externalData map[string]any) (int64, error)
}

type TenantRetriever interface {
	GetTenant(ctx context.Context, tenantID int64) (tenants.Tenant, error)
}

type CourseInfoRetriever interface {
	GetCourseInfoFromLTILaunch(ctx context.Context, tenantID int64, claims jwt.MapClaims) (string, map[string]any, error)
}

type ProcessLaunchCommand struct {
	IDToken string
	State   string
}

type ProcessLaunchResult struct {
	TargetLinkURI string
	UserID      int64
	CourseID    int64
	TenantID    int64
}

func NewProcessLaunchUseCase(
	ltiSessionRepository domain.LTISessionRepository,
	registrationRepository domain.RegistrationRepository,
	ltiUserLinkRepository domain.LTIUserLinkRepository,
	ltiCourseLinkRepository domain.LTICourseLinkRepository,
	
	tenantRetriever TenantRetriever,
	idTokenVerifier IDTokenVerifier,
	userCreator UserCreator,
	courseCreator CourseCreator,
	courseInfoRetriever CourseInfoRetriever,
) *ProcessLaunchUseCase {
	return &ProcessLaunchUseCase{
		ltiSessionRepository:    	ltiSessionRepository,
		registrationRepository:  	registrationRepository,
		ltiUserLinkRepository:   	ltiUserLinkRepository,
		ltiCourseLinkRepository: 	ltiCourseLinkRepository,

		tenantRetriever:          tenantRetriever,
		idTokenVerifier:         	idTokenVerifier,
		userCreator:             	userCreator,
		courseCreator:           	courseCreator,
		courseInfoRetriever:     	courseInfoRetriever,
	}
}

func (u *ProcessLaunchUseCase) Execute(ctx context.Context, cmd ProcessLaunchCommand) (ProcessLaunchResult, error) {
	session, err := u.loadSession(ctx, cmd.State)
	if err != nil {
		return ProcessLaunchResult{}, err
	}

	registration, err := u.loadRegistration(ctx, session.Issuer(), session.ClientID())
	if err != nil {
		return ProcessLaunchResult{}, err
	}

	claims, err := u.validateIDToken(ctx, cmd.IDToken, session, registration)
	if err != nil {
		return ProcessLaunchResult{}, err
	}

	tenant, err := u.tenantRetriever.GetTenant(ctx, registration.TenantID)
	if err != nil {
		return ProcessLaunchResult{}, err
	}

	userID, err := u.resolveUser(ctx, claims, session)
	if err != nil {
		return ProcessLaunchResult{}, err
	}

	courseID, err := u.resolveCourse(ctx, claims, tenant)
	if err != nil {
		return ProcessLaunchResult{}, err
	}

	return ProcessLaunchResult{
		TargetLinkURI: session.TargetLinkURI(),
		UserID:      userID,
		CourseID:    courseID,
		TenantID:    session.TenantID(),
	}, nil
}

// `loadSession` retrieves the LTI session from the repository using the provided
// state. It returns an error if the session is not found or has expired.
func (u *ProcessLaunchUseCase) loadSession(ctx context.Context, state string) (*domain.LTISession, error) {
	session, err := u.ltiSessionRepository.GetByState(ctx, state)
	if err != nil {
		return nil, err
	}

	if session == nil || session.IsExpired() {
		return nil, apperr.New(
			apperr.CodeNotFound,
			"session_not_found",
			"LTI session not found or expired",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
		)
	}

	return session, nil
}

// `loadRegistration` retrieves the LTI registration from the repository using the
// provided issuer and client ID. It returns an error if the registration is not found.
func (u *ProcessLaunchUseCase) loadRegistration(ctx context.Context, issuer string, clientID string) (*domain.Registration, error) {
	registration, err := u.registrationRepository.GetByIssuerAndClientID(ctx, issuer, clientID)
	if err != nil {
		return nil, err
	}
	if registration == nil {
		return nil, apperr.New(
			apperr.CodeNotFound,
			"registration_not_found",
			"LTI registration not found for issuer and client ID",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
		)
	}
	return registration, nil
}

// `validateIDToken` validates the provided ID token against the LTI session and
// registration. It checks the token's signature, claims, and ensures that it
// matches the session's expected values. If validation is successful, it returns
// the parsed claims.
func (u *ProcessLaunchUseCase) validateIDToken(ctx context.Context, idToken string, session *domain.LTISession, registration *domain.Registration) (jwt.MapClaims, error) {
	claims, err := u.idTokenVerifier.Verify(ctx, idToken, registration.JWKEndpoint)
	if err != nil {
		if verificationErr, ok := err.(*IDTokenVerificationError); ok {
			switch verificationErr.Code {
			case IDTokenVerificationParseError:
				return nil, apperr.New(
					apperr.CodeInternal,
					"invalid_id_token",
					"Failed to parse ID token",
					apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
					apperr.WithCause(err),
				)
			case IDTokenVerificationMissingKID:
				return nil, apperr.New(
					apperr.CodeInternal,
					"invalid_id_token",
					"ID token missing 'kid' header",
					apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
				)
			case IDTokenVerificationJWKSClientError:
				return nil, apperr.New(
					apperr.CodeInternal,
					"jwks_client_error",
					"Failed to create JWKS client",
					apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
					apperr.WithCause(err),
				)
			case IDTokenVerificationValidationError:
				return nil, apperr.New(
					apperr.CodeInternal,
					"invalid_id_token",
					"ID token validation failed",
					apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
					apperr.WithCause(err),
				)
			case IDTokenVerificationInvalidClaims:
				return nil, apperr.New(
					apperr.CodeInternal,
					"invalid_id_token",
					"Invalid ID token claims",
					apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
				)
			}
		}

		return nil, apperr.New(
			apperr.CodeInternal,
			"invalid_id_token",
			"ID token validation failed",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
			apperr.WithCause(err),
		)
	}

	if err := validateLaunchClaims(claims, session); err != nil {
		return nil, err
	}

	if err := u.ltiSessionRepository.Delete(ctx, session.State()); err != nil {
		return nil, err
	}

	return claims, nil
}

// `resolveUser` checks if a user link exists for the given LTI session and
// claims. If a link exists, it returns the associated user ID. If not, it creates
// a new user and establishes a link between the LTI user and the newly created
// user, returning the new user ID.
func (u *ProcessLaunchUseCase) resolveUser(ctx context.Context, claims jwt.MapClaims, session *domain.LTISession) (int64, error) {
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return 0, apperr.New(
			apperr.CodeValidation,
			"missing_sub_claim",
			"ID token missing 'sub' claim",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
		)
	}

	userLink, err := u.ltiUserLinkRepository.GetBySubAndIssuer(ctx, sub, session.Issuer())
	if err != nil {
		return 0, err
	}
	if userLink == nil {
		name, ok := claims["name"].(string)
		if !ok || name == "" {
			return 0, apperr.New(
				apperr.CodeValidation,
				"missing_name_claim",
				"ID token missing 'name' claim",
				apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
			)
		}

		userID, err := u.userCreator.CreateUser(ctx, sub, name)
		if err != nil {
			return 0, err
		}
		userLink = domain.NewLTIUserLink(sub, session.Issuer(), userID)
		if err := u.ltiUserLinkRepository.Create(ctx, userLink); err != nil {
			return 0, err
		}
	}

	return userLink.UserID(), nil
}


// `resolveCourse` checks if a course link exists for the given tenant and 
// context. If no link exists, a new course is created. The pair 
// (`tenantID`, `contextID`) is assumed to be unique. This is not guaranteed
// in LTI 1.3. However, it is recommended in the spec, and no other
// one-to-one mapping from a set of claims to a course seems to exist.
// https://www.imsglobal.org/spec/lti/v1p3#context-claim
func (u *ProcessLaunchUseCase) resolveCourse(ctx context.Context, claims jwt.MapClaims, tenant tenants.Tenant) (int64, error) {
	contextClaim, ok := claims["https://purl.imsglobal.org/spec/lti/claim/context"].(map[string]any)
	if !ok {
		return 0, apperr.New(
			apperr.CodeValidation,
			"missing_context_claim",
			"ID token missing 'context' claim",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
		)
	}

	contextID, ok := contextClaim["id"].(string)
	if !ok || contextID == "" {
		return 0, apperr.New(
			apperr.CodeValidation,
			"missing_context_id",
			"ID token context claim missing 'id'",
			apperr.WithOp("lti.application.ProcessLaunchUseCase.Execute"),
		)
	}

	courseLink, err := u.ltiCourseLinkRepository.GetByTenantAndContext(ctx, tenant.ID, contextID)
	if err != nil {
		return 0, err
	}
	if courseLink == (domain.LTICourseLink{}) {
		// No course associated with this request was found.
		// Instead of returning an error, we create a new course.

		courseTitle, ok := contextClaim["title"].(string)
		if !ok {
			courseTitle = "Untitled Course"
		}

		externalCourseID, externalCourseInfo, err := u.courseInfoRetriever.GetCourseInfoFromLTILaunch(ctx, tenant.ID,claims)
		if err != nil {
			return 0, err
		}

		courseID, err := u.courseCreator.CreateCourse(ctx, courseTitle, tenant.ID, externalCourseID, externalCourseInfo)
		if err != nil {
			return 0, err
		}

		courseLink = domain.NewLTICourseLink(tenant.ID, contextID, courseID)
		if err = u.ltiCourseLinkRepository.Create(ctx, courseLink); err != nil {
			return 0, err
		}

	}

	return courseLink.CourseID(), nil

}


// `validateLaunchClaims` checks that the claims in the ID token match the 
// expected values in the LTI session. It ensures that the `iss`, `nonce`,
// and `aud` claims are consistent with the session.
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

	if targetLinkURI, _ := claims["https://purl.imsglobal.org/spec/lti/claim/target_link_uri"].(string); targetLinkURI != session.TargetLinkURI() {
		return apperr.New(
			apperr.CodeValidation,
			"invalid_id_token",
			"ID token 'target_link_uri' claim does not match session target link URI",
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
