package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rewritetest/internal/auth"
	"rewritetest/internal/lti/internal/application"
	"rewritetest/internal/lti/internal/domain"
	sharedmiddleware "rewritetest/internal/shared/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

var _ SessionCreator = (*fakeHandlerSessionCreator)(nil)
var _ domain.RegistrationRepository = (*fakeHandlerRegistrationRepo)(nil)
var _ domain.LTISessionRepository = (*fakeHandlerSessionRepo)(nil)
var _ application.IDTokenVerifier = (*fakeHandlerVerifier)(nil)
var _ domain.LTIUserLinkRepository = (*fakeHandlerUserLinkRepo)(nil)
var _ domain.LTICourseLinkRepository = (*fakeHandlerCourseLinkRepo)(nil)
var _ application.UserCreator = (*fakeHandlerUserCreator)(nil)
var _ application.CourseCreator = (*fakeHandlerCourseCreator)(nil)


type fakeHandlerSessionCreator struct {
	session auth.Session
	err     error
}

func (f *fakeHandlerSessionCreator) CreateSession(context.Context, int64, int64) (auth.Session, error) {
	if f.err != nil {
		return auth.Session{}, f.err
	}
	return f.session, nil
}


type fakeHandlerRegistrationRepo struct {
	registration *domain.Registration
}

func (f *fakeHandlerRegistrationRepo) Create(context.Context, domain.Registration) error { return nil }
func (f *fakeHandlerRegistrationRepo) Save(context.Context, domain.Registration) error   { return nil }
func (f *fakeHandlerRegistrationRepo) GetByIssuerAndClientID(context.Context, string, string) (*domain.Registration, error) {
	return f.registration, nil
}


type fakeHandlerSessionRepo struct {
	session *domain.LTISession
	created *domain.LTISession
}

func (f *fakeHandlerSessionRepo) Create(_ context.Context, session *domain.LTISession) error {
	f.created = session
	return nil
}
func (f *fakeHandlerSessionRepo) GetByState(context.Context, string) (*domain.LTISession, error) {
	return f.session, nil
}
func (f *fakeHandlerSessionRepo) Delete(context.Context, string) error { return nil }


type fakeHandlerVerifier struct {
	claims jwt.MapClaims
	err    error
}

func (f *fakeHandlerVerifier) Verify(context.Context, string, string) (jwt.MapClaims, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.claims, nil
}


type fakeHandlerUserLinkRepo struct {
	link *domain.LTIUserLink
}

func (f *fakeHandlerUserLinkRepo) GetBySubAndIssuer(context.Context, string, string) (*domain.LTIUserLink, error) {
	return f.link, nil
}
func (f *fakeHandlerUserLinkRepo) Create(context.Context, *domain.LTIUserLink) error { return nil }


type fakeHandlerCourseLinkRepo struct {
	link domain.LTICourseLink
}

func (f *fakeHandlerCourseLinkRepo) GetByTenantAndContext(context.Context, int64, string) (domain.LTICourseLink, error) {
	return f.link, nil
}
func (f *fakeHandlerCourseLinkRepo) Create(context.Context, domain.LTICourseLink) error { return nil }


type fakeHandlerUserCreator struct{}

func (fakeHandlerUserCreator) CreateUser(context.Context, string, string) (int64, error) {
	return 1, nil
}


type fakeHandlerCourseCreator struct{}

func (fakeHandlerCourseCreator) CreateCourse(context.Context, string, int64) (int64, error) {
	return 2, nil
}


func TestHandleLoginInitiation_InvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&fakeHandlerSessionCreator{}, buildGetLaunchRedirectUseCase(), buildProcessLaunchUseCase(), buildBeginAuthenticationUseCase(), "http://tool.example/lti/auth/check")
	r := gin.New()
	r.Use(sharedmiddleware.ErrorHandler())
	r.POST("/lti/authorize", h.handleLoginInitiation)

	req := httptest.NewRequest(http.MethodPost, "/lti/authorize", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), `"code":"VALIDATION"`)
}

func TestHandleLaunch_PassesThroughAppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&fakeHandlerSessionCreator{}, buildGetLaunchRedirectUseCase(), buildProcessLaunchUseCaseMissingSession(), buildBeginAuthenticationUseCase(), "http://tool.example/lti/auth/check")
	r := gin.New()
	r.Use(sharedmiddleware.ErrorHandler())
	r.POST("/lti/authorize/check", h.handleLaunch)

	req := httptest.NewRequest(http.MethodPost, "/lti/authorize/check", strings.NewReader("id_token=x&state=state-1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), `"code":"NOT_FOUND"`)
}

func TestHandleLaunch_SuccessSetsCookieAndRedirects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(
		&fakeHandlerSessionCreator{session: auth.Session{ID: "sess-1", ExpiresAt: time.Now().Add(time.Hour)}},
		buildGetLaunchRedirectUseCase(),
		buildProcessLaunchUseCase(),
		buildBeginAuthenticationUseCase(),
		"http://tool.example/lti/auth/check",
	)
	r := gin.New()
	r.Use(sharedmiddleware.ErrorHandler())
	r.POST("/lti/authorize/check", h.handleLaunch)

	req := httptest.NewRequest(http.MethodPost, "/lti/authorize/check", strings.NewReader("id_token=x&state=state-1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "https://tool.example/app", rec.Header().Get("Location"))

	var authCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "AUTH_TOKEN" {
			authCookie = c
			break
		}
	}
	require.NotNil(t, authCookie)
	require.Equal(t, "sess-1", authCookie.Value)
	require.True(t, authCookie.HttpOnly)
}

func buildGetLaunchRedirectUseCase() *application.GetLaunchRedirectUseCase {
	return application.NewGetLaunchRedirectUseCase(
		&fakeHandlerRegistrationRepo{registration: domain.NewRegistration(
			"https://issuer.example",
			"client-1",
			"https://issuer.example/auth",
			"https://issuer.example/jwks",
			"https://issuer.example/service-auth",
			"https://issuer.example/logout",
			7,
		)},
		&fakeHandlerSessionRepo{},
	)
}

func buildProcessLaunchUseCase() *application.ProcessLaunchUseCase {
	return application.NewProcessLaunchUseCase(
		&fakeHandlerSessionRepo{session: domain.NewLTISession(
			"state-1",
			"nonce-1",
			"https://issuer.example",
			"client-1",
			"https://tool.example/app",
			7,
			time.Now().Add(-time.Minute),
			time.Now().Add(time.Hour),
		)},
		&fakeHandlerRegistrationRepo{registration: domain.NewRegistration(
			"https://issuer.example",
			"client-1",
			"https://issuer.example/auth",
			"https://issuer.example/jwks",
			"https://issuer.example/service-auth",
			"https://issuer.example/logout",
			7,
		)},
		&fakeHandlerUserLinkRepo{link: domain.NewLTIUserLink("sub-1", "https://issuer.example", 44)},
		&fakeHandlerCourseLinkRepo{link: domain.NewLTICourseLink(7, "ctx-1", 99)},
		&fakeHandlerVerifier{claims: jwt.MapClaims{
			"iss":   "https://issuer.example",
			"nonce": "nonce-1",
			"aud":   "client-1",
			"sub":   "sub-1",
			"name":  "Alice",
			"https://purl.imsglobal.org/spec/lti/claim/context": map[string]any{
				"id": "ctx-1",
			},
		}},
		fakeHandlerUserCreator{},
		fakeHandlerCourseCreator{},
	)
}

func buildProcessLaunchUseCaseMissingSession() *application.ProcessLaunchUseCase {
	return application.NewProcessLaunchUseCase(
		&fakeHandlerSessionRepo{session: nil},
		&fakeHandlerRegistrationRepo{registration: nil},
		&fakeHandlerUserLinkRepo{},
		&fakeHandlerCourseLinkRepo{},
		&fakeHandlerVerifier{},
		fakeHandlerUserCreator{},
		fakeHandlerCourseCreator{},
	)
}

func buildBeginAuthenticationUseCase() *application.BeginAuthenticationUseCase {
	return application.NewBeginAuthenticationUseCase(
		nil,
	)
}