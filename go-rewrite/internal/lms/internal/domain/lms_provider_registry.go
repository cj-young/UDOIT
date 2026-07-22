package domain

import "context"

type LMSProviderRegistry interface {
	Get(ctx context.Context, lmsKey LMSType) (LMSProvider, error)
	RegisterProvider(lmsKey LMSType, provider LMSProvider)
}