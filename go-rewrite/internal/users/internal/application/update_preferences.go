package application

import (
	"context"
	"log/slog"

	"rewritetest/internal/shared/apperr"
	sharedAuth "rewritetest/internal/shared/auth"
	"rewritetest/internal/users/internal/domain"
)

type UpdatePreferencesCommand struct {
	UserID       int64
	Theme        *string
	TextSpacing  *int
	FontSize     *string
	FontFamily   *string
	AlertTimeout *int
	Language     *string
}

type I18NService interface {
	GetLabels(ctx context.Context, lang string) (map[string]string, error)
}

type UpdatePreferencesUseCase struct {
	userRepository domain.UserRepository
	i18nService    I18NService
}

func NewUpdatePreferencesUseCase(userRepository domain.UserRepository, i18nService I18NService) *UpdatePreferencesUseCase {
	return &UpdatePreferencesUseCase{
		userRepository: userRepository,
		i18nService:    i18nService,
	}
}

func (u *UpdatePreferencesUseCase) Execute(ctx context.Context, principal sharedAuth.Principal, cmd UpdatePreferencesCommand) (map[string]string, error) {
	if principal.AgentID != cmd.UserID {
		return nil, apperr.New(
			apperr.CodeUnauthorized, "forbidden", "You do not have permission to update these preferences",
			apperr.WithOp("users.application.update_preferences.Execute"),
		)
	}

	user, err := u.userRepository.GetByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, apperr.New(
			apperr.CodeNotFound, "user_not_found", "The requested user was not found",
			apperr.WithOp("users.application.update_preferences.Execute"),
		)
	}

	slog.Info("Updating preferences for user", "user", user)

	preferences := user.Preferences()
	update, err := toPreferencesUpdate(cmd)
	if err != nil {
		return nil, err
	}
	newPreferences := preferences.ApplyChanges(update)

	if newPreferences == preferences {
		return nil, nil
	}

	user.SetPreferences(newPreferences)

	languageChanged := update.Language != nil && (preferences.Language == nil || *update.Language != *preferences.Language)

	var labels map[string]string
	if languageChanged {
		l, err := u.i18nService.GetLabels(ctx, *update.Language)
		if err != nil {
			return nil, apperr.New(
				apperr.CodeInternal, "failed_to_fetch_labels", "An error occurred while fetching labels for the new language",
				apperr.WithOp("users.application.update_preferences.Execute"),
			)
		}
		labels = l
	}

	err = u.userRepository.Update(ctx, user)
	if err != nil {
		slog.Error("Failed to update user preferences", "error", err)
		return nil, apperr.New(
			apperr.CodeInternal, "failed_to_save_user", "An error occurred while saving the user",
			apperr.WithOp("users.application.update_preferences.Execute"),
		)
	}

	return labels, nil
}

func toPreferencesUpdate(cmd UpdatePreferencesCommand) (domain.PreferencesUpdate, error) {
	update := domain.PreferencesUpdate{}

	if cmd.Theme != nil {
		theme := domain.Theme(*cmd.Theme)

		if !theme.IsValid() {
			return update, apperr.New(
				apperr.CodeValidation, "invalid_theme", "The provided theme is invalid",
				apperr.WithOp("users.application.update_preferences.toPreferencesUpdate"),
			)
		}

		update.Theme = &theme
	}

	update.TextSpacing = cmd.TextSpacing
	update.FontSize = cmd.FontSize
	update.FontFamily = cmd.FontFamily
	update.AlertTimeout = cmd.AlertTimeout
	update.Language = cmd.Language

	return update, nil
}
