package domain

import (
	"context"
	"rewritetest/internal/shared/auth"

	"github.com/golang-jwt/jwt/v5"
)

// `FullLMSProvider` is split up into multiple smaller interfaces.
// With the current implementation, this doesn't provide any functionality,
// but it lays groundwork for better interface segregation and
type FullLMSProvider interface {
	FileProvider
	AuthenticationProvider
	ConfigValidator
	ScanProvider
	LTIProvider
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

type LMSCourse struct {
	ID						int64
	ExternalID		string
	ExternalData	map[string]any
}

type LMSContent struct {
	ID						int64
	ExternalID		string
	ExternalData	map[string]any
	HTML					string
}

type ScanProvider interface {
	// The current course content is sent to the LMS provider to allow it to
	// skip fetching content that is already up to date.
	// If new internal content items are created out of the return, the LMS provider
	// expects them to be explicitly registered later.
	GetContent(ctx context.Context, tenantConfig LMSProviderConfig, course LMSCourse, currentContent []LMSContent, userID int64) ([]CourseContent, error)
}

type LTIProvider interface {
	GetCourseInfoFromLTILaunch(ctx context.Context, tenantConfig LMSProviderConfig, claims jwt.MapClaims) (string, map[string]any, error)
}

