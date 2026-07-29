package domain

import "context"

type OAuthRedirectProcessor interface {
	ProcessOAuthRedirect(ctx context.Context, config LMSProviderConfig, authAttempt AuthAttempt, code string) (string, error)
}