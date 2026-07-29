package domain

import "context"


type LMSProviderRegistry interface {
	// Get retrieves the LMS provider associated with the given LMS key and
	// errors if no provider is found for the given LMS key.
	Get(ctx context.Context, lmsKey LMSType) (FullLMSProvider, error)
	RegisterProvider(lmsKey LMSType, provider FullLMSProvider)
}