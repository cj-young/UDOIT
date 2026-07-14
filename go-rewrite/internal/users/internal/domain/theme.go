package domain

import "rewritetest/internal/shared/apperr"

type Theme string

const (
	ThemeLight Theme = "light"
	ThemeDark  Theme = "dark"
)

func NewTheme(value string) (*Theme, error) {
	t := Theme(value)
	if !t.IsValid() {
		return nil, apperr.New(
			apperr.CodeValidation, "invalid_theme", "The provided theme is invalid",
			apperr.WithOp("users.domain.NewTheme"),
		)
	}

	return &t, nil
}

func (t Theme) IsValid() bool {
	switch t {
	case ThemeLight, ThemeDark:
		return true
	default:
		return false
	}
}
