package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"rewritetest/internal/lti/internal/domain"
	"rewritetest/internal/shared/apperr"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

var _ domain.LTISessionRepository = (*fakeProcessSessionRepo)(nil)
var _ domain.RegistrationRepository = (*fakeProcessRegistrationRepo)(nil)
var _ domain.LTIUserLinkRepository = (*fakeProcessUserLinkRepo)(nil)
var _ domain.LTICourseLinkRepository = (*fakeProcessCourseLinkRepo)(nil)
var _ IDTokenVerifier = (*fakeProcessVerifier)(nil)
var _ UserCreator = (*fakeProcessUserCreator)(nil)
var _ CourseCreator = (*fakeProcessCourseCreator)(nil)

type fakeProcessSessionRepo struct {
	session      *domain.LTISession
	getErr       error
	deletedState string
}

func (f *fakeProcessSessionRepo) Create(context.Context, *domain.LTISession) error { return nil }

func (f *fakeProcessSessionRepo) GetByState(context.Context, string) (*domain.LTISession, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.session, nil
}

func (f *fakeProcessSessionRepo) Delete(_ context.Context, state string) error {
	f.deletedState = state
	return nil
}

type fakeProcessRegistrationRepo struct {
	registration *domain.Registration
}

func (f *fakeProcessRegistrationRepo) Create(context.Context, domain.Registration) error { return nil }
func (f *fakeProcessRegistrationRepo) Save(context.Context, domain.Registration) error   { return nil }
func (f *fakeProcessRegistrationRepo) GetByIssuerAndClientID(context.Context, string, string) (*domain.Registration, error) {
	return f.registration, nil
}

type fakeProcessUserLinkRepo struct {
	link    *domain.LTIUserLink
	created *domain.LTIUserLink
}

func (f *fakeProcessUserLinkRepo) GetBySubAndIssuer(context.Context, string, string) (*domain.LTIUserLink, error) {
	return f.link, nil
}

func (f *fakeProcessUserLinkRepo) Create(_ context.Context, link *domain.LTIUserLink) error {
	f.created = link
	return nil
}

type fakeProcessCourseLinkRepo struct {
	link    domain.LTICourseLink
	created domain.LTICourseLink
}

func (f *fakeProcessCourseLinkRepo) Create(_ context.Context, link domain.LTICourseLink) error {
	f.created = link
	return nil
}

func (f *fakeProcessCourseLinkRepo) GetByTenantAndContext(context.Context, int64, string) (domain.LTICourseLink, error) {
	return f.link, nil
}

type fakeProcessVerifier struct {
	claims jwt.MapClaims
	err    error
}

func (f *fakeProcessVerifier) Verify(context.Context, string, string) (jwt.MapClaims, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.claims, nil
}

type fakeProcessUserCreator struct {
	id     int64
	called bool
}

func (f *fakeProcessUserCreator) CreateUser(context.Context, string, string) (int64, error) {
	f.called = true
	return f.id, nil
}

type fakeProcessCourseCreator struct {
	id     int64
	called bool
}

func (f *fakeProcessCourseCreator) CreateCourse(context.Context, string, int64) (int64, error) {
	f.called = true
	return f.id, nil
}

func TestProcessLaunchExecute_SessionMissing(t *testing.T) {
	uc := NewProcessLaunchUseCase(
		&fakeProcessSessionRepo{},
		&fakeProcessRegistrationRepo{registration: validRegistration()},
		&fakeProcessUserLinkRepo{},
		&fakeProcessCourseLinkRepo{},
		&fakeProcessVerifier{},
		&fakeProcessUserCreator{id: 10},
		&fakeProcessCourseCreator{id: 20},
	)

	_, err := uc.Execute(context.Background(), ProcessLaunchCommand{IDToken: "token", State: "state-1"})

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeNotFound))
}

func TestProcessLaunchExecute_VerifierErrorMapped(t *testing.T) {
	sessionRepo := &fakeProcessSessionRepo{session: validSession()}
	uc := NewProcessLaunchUseCase(
		sessionRepo,
		&fakeProcessRegistrationRepo{registration: validRegistration()},
		&fakeProcessUserLinkRepo{},
		&fakeProcessCourseLinkRepo{},
		&fakeProcessVerifier{err: NewIDTokenVerificationError(IDTokenVerificationParseError, errors.New("bad token"))},
		&fakeProcessUserCreator{id: 10},
		&fakeProcessCourseCreator{id: 20},
	)

	_, err := uc.Execute(context.Background(), ProcessLaunchCommand{IDToken: "bad", State: "state-1"})

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeInternal))
	require.Empty(t, sessionRepo.deletedState)
}

func TestProcessLaunchExecute_ExistingLinksPath(t *testing.T) {
	sessionRepo := &fakeProcessSessionRepo{session: validSession()}
	userCreator := &fakeProcessUserCreator{id: 10}
	courseCreator := &fakeProcessCourseCreator{id: 20}
	uc := NewProcessLaunchUseCase(
		sessionRepo,
		&fakeProcessRegistrationRepo{registration: validRegistration()},
		&fakeProcessUserLinkRepo{link: domain.NewLTIUserLink("sub-1", "https://issuer.example", 77)},
		&fakeProcessCourseLinkRepo{link: domain.NewLTICourseLink(7, "ctx-1", 88)},
		&fakeProcessVerifier{claims: validClaims()},
		userCreator,
		courseCreator,
	)

	result, err := uc.Execute(context.Background(), ProcessLaunchCommand{IDToken: "token", State: "state-1"})

	require.NoError(t, err)
	require.Equal(t, int64(77), result.UserID)
	require.Equal(t, int64(88), result.CourseID)
	require.Equal(t, int64(7), result.TenantID)
	require.Equal(t, "https://tool.example/app", result.RedirectURL)
	require.Equal(t, "state-1", sessionRepo.deletedState)
	require.False(t, userCreator.called)
	require.False(t, courseCreator.called)
}

func TestProcessLaunchExecute_CreatesLinksWhenMissing(t *testing.T) {
	sessionRepo := &fakeProcessSessionRepo{session: validSession()}
	userLinkRepo := &fakeProcessUserLinkRepo{}
	courseLinkRepo := &fakeProcessCourseLinkRepo{}
	userCreator := &fakeProcessUserCreator{id: 55}
	courseCreator := &fakeProcessCourseCreator{id: 66}
	uc := NewProcessLaunchUseCase(
		sessionRepo,
		&fakeProcessRegistrationRepo{registration: validRegistration()},
		userLinkRepo,
		courseLinkRepo,
		&fakeProcessVerifier{claims: validClaims()},
		userCreator,
		courseCreator,
	)

	result, err := uc.Execute(context.Background(), ProcessLaunchCommand{IDToken: "token", State: "state-1"})

	require.NoError(t, err)
	require.Equal(t, int64(55), result.UserID)
	require.Equal(t, int64(66), result.CourseID)
	require.True(t, userCreator.called)
	require.True(t, courseCreator.called)
	require.NotNil(t, userLinkRepo.created)
	require.Equal(t, int64(55), userLinkRepo.created.UserID())
	require.Equal(t, int64(7), courseLinkRepo.created.TenantID())
	require.Equal(t, "ctx-1", courseLinkRepo.created.ContextID())
	require.Equal(t, int64(66), courseLinkRepo.created.CourseID())
}

func validSession() *domain.LTISession {
	return domain.NewLTISession(
		"state-1",
		"nonce-1",
		"https://issuer.example",
		"client-1",
		"https://tool.example/app",
		7,
		time.Now().Add(-1*time.Minute),
		time.Now().Add(10*time.Minute),
	)
}

func validRegistration() *domain.Registration {
	return domain.NewRegistration(
		"https://issuer.example",
		"client-1",
		"https://issuer.example/auth",
		"https://issuer.example/jwks",
		"https://issuer.example/service-auth",
		"https://issuer.example/logout",
		7,
	)
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss":   "https://issuer.example",
		"nonce": "nonce-1",
		"aud":   "client-1",
		"sub":   "sub-1",
		"name":  "Alice",
		"https://purl.imsglobal.org/spec/lti/claim/context": map[string]any{
			"id":    "ctx-1",
			"title": "Biology",
		},
	}
}
