package application

import (
	"context"
	"errors"
	"testing"

	"rewritetest/internal/shared/apperr"
	sharedauth "rewritetest/internal/shared/auth"
	"rewritetest/internal/users/internal/domain"

	"github.com/stretchr/testify/require"
)

type fakeUserRepository struct {
	getByIDFn  func(ctx context.Context, id int64) (*domain.User, error)
	updateFn   func(ctx context.Context, user *domain.User) error
	updateCall int
}

func (f *fakeUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return f.getByIDFn(ctx, id)
}

func (f *fakeUserRepository) Create(context.Context, *domain.User) error {
	return nil
}

func (f *fakeUserRepository) Update(ctx context.Context, user *domain.User) error {
	f.updateCall++
	if f.updateFn != nil {
		return f.updateFn(ctx, user)
	}
	return nil
}

type fakeI18NService struct {
	labels   map[string]string
	err      error
	callLang string
	called   bool
}

func (f *fakeI18NService) GetLabels(_ context.Context, lang string) (map[string]string, error) {
	f.called = true
	f.callLang = lang
	if f.err != nil {
		return nil, f.err
	}
	return f.labels, nil
}

func TestUpdatePreferencesUseCase_Unauthorized(t *testing.T) {
	repo := &fakeUserRepository{}
	i18n := &fakeI18NService{}
	uc := NewUpdatePreferencesUseCase(repo, i18n)

	_, err := uc.Execute(context.Background(), sharedauth.Principal{AgentID: 10}, UpdatePreferencesCommand{UserID: 11})

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeUnauthorized))
	require.Equal(t, 0, repo.updateCall)
	require.False(t, i18n.called)
}

func TestUpdatePreferencesUseCase_InvalidTheme(t *testing.T) {
	user := domain.RehydrateUser(10, "user", "User", domain.Preferences{})
	repo := &fakeUserRepository{
		getByIDFn: func(_ context.Context, id int64) (*domain.User, error) {
			require.Equal(t, int64(10), id)
			return &user, nil
		},
	}
	i18n := &fakeI18NService{}
	uc := NewUpdatePreferencesUseCase(repo, i18n)

	invalid := "sepia"
	_, err := uc.Execute(context.Background(), sharedauth.Principal{AgentID: 10}, UpdatePreferencesCommand{UserID: 10, Theme: &invalid})

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeValidation))
	require.Equal(t, 0, repo.updateCall)
	require.False(t, i18n.called)
}

func TestUpdatePreferencesUseCase_NoChangesSkipsUpdate(t *testing.T) {
	theme := domain.ThemeLight
	lang := "en"
	spacing := 1
	user := domain.RehydrateUser(10, "user", "User", domain.Preferences{Theme: &theme, Language: &lang, TextSpacing: &spacing})
	repo := &fakeUserRepository{
		getByIDFn: func(_ context.Context, _ int64) (*domain.User, error) { return &user, nil },
	}
	i18n := &fakeI18NService{}
	uc := NewUpdatePreferencesUseCase(repo, i18n)

	labels, err := uc.Execute(context.Background(), sharedauth.Principal{AgentID: 10}, UpdatePreferencesCommand{UserID: 10})

	require.NoError(t, err)
	require.Nil(t, labels)
	require.Equal(t, 0, repo.updateCall)
	require.False(t, i18n.called)
}

func TestUpdatePreferencesUseCase_LanguageChangeFetchesLabelsAndUpdates(t *testing.T) {
	oldLang := "en"
	theme := domain.ThemeLight
	user := domain.RehydrateUser(10, "user", "User", domain.Preferences{Theme: &theme, Language: &oldLang})
	repo := &fakeUserRepository{
		getByIDFn: func(_ context.Context, _ int64) (*domain.User, error) { return &user, nil },
	}
	i18n := &fakeI18NService{labels: map[string]string{"hello": "hola"}}
	uc := NewUpdatePreferencesUseCase(repo, i18n)

	labels, err := uc.Execute(context.Background(), sharedauth.Principal{AgentID: 10}, UpdatePreferencesCommand{UserID: 10, Language: strPtr("es")})

	require.NoError(t, err)
	require.Equal(t, map[string]string{"hello": "hola"}, labels)
	require.True(t, i18n.called)
	require.Equal(t, "es", i18n.callLang)
	require.Equal(t, 1, repo.updateCall)
	require.NotNil(t, user.Preferences().Language)
	require.Equal(t, "es", *user.Preferences().Language)
}

func TestUpdatePreferencesUseCase_SaveFailureIsInternal(t *testing.T) {
	user := domain.RehydrateUser(10, "user", "User", domain.Preferences{})
	repo := &fakeUserRepository{
		getByIDFn: func(_ context.Context, _ int64) (*domain.User, error) { return &user, nil },
		updateFn:  func(_ context.Context, _ *domain.User) error { return errors.New("db down") },
	}
	i18n := &fakeI18NService{}
	uc := NewUpdatePreferencesUseCase(repo, i18n)

	_, err := uc.Execute(context.Background(), sharedauth.Principal{AgentID: 10}, UpdatePreferencesCommand{UserID: 10, Language: strPtr("fr")})

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeInternal))
	require.Equal(t, 1, repo.updateCall)
}

func strPtr(v string) *string { return &v }
func intPtr(v int) *int       { return &v }
