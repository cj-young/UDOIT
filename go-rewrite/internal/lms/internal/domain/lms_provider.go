package domain

import (
	"context"
	"rewritetest/internal/shared/auth"
)

type FullLMSProvider interface {
	FileProvider
	AuthenticationProvider
	ConfigValidator
}

type LMSFile struct {
	ID 						int64
	ExternalID 		string
	ExternalData 	map[string]any
}
type FileProvider interface {
	DeleteFile(ctx context.Context, principal auth.Principal, config LMSProviderConfig, file LMSFile) error
}

type AuthenticationProvider interface {
	BeginAuthentication(ctx context.Context, config LMSProviderConfig, userID int64, targetLinkURI string) (AuthChallenge, error)
}

type ConfigValidator interface {
	ValidateConfig(configData map[string]any) error
}

