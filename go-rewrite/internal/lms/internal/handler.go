package internal

import (
	"net/http"
	"strings"
	"time"

	"rewritetest/internal/lms/internal/application"
	"rewritetest/internal/shared/apperr"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	syncSessionUseCase *application.SyncSessionUseCase
	internalSyncSecret string
}

type SyncSessionRequest struct {
	UserID            int64          `json:"user_id" binding:"required"`
	LMSKey            string         `json:"lms_key" binding:"required"`
	ExternalUserID    string         `json:"external_user_id"`
	APIDomain         string         `json:"api_domain"`
	Metadata          map[string]any `json:"metadata"`
	CredentialSchema  string         `json:"credential_schema"`
	CredentialPayload map[string]any `json:"credential_payload"`
	CredentialExpires string         `json:"credential_expires_at"`
}

func NewHandler(syncSessionUseCase *application.SyncSessionUseCase, internalSyncSecret string) *Handler {
	return &Handler{
		syncSessionUseCase: syncSessionUseCase,
		internalSyncSecret: internalSyncSecret,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/sync-session", h.handleSyncSession)
}

func (h *Handler) handleSyncSession(c *gin.Context) {
	if !h.isAuthorized(c.GetHeader("X-Internal-Sync-Secret")) {
		c.Error(apperr.New(
			apperr.CodeUnauthorized,
			"unauthorized_sync",
			"Unauthorized internal sync request",
			apperr.WithOp("lms.handler.handleSyncSession"),
		))
		return
	}

	var req SyncSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.New(
			apperr.CodeValidation,
			"invalid_request_payload",
			"Invalid request payload",
			apperr.WithOp("lms.handler.handleSyncSession"),
			apperr.WithCause(err),
		))
		return
	}

	var credentialExpires *time.Time
	if strings.TrimSpace(req.CredentialExpires) != "" {
		parsed, err := time.Parse(time.RFC3339, req.CredentialExpires)
		if err != nil {
			c.Error(apperr.New(
				apperr.CodeValidation,
				"invalid_credential_expiration",
				"credential_expires_at must be RFC3339",
				apperr.WithOp("lms.handler.handleSyncSession"),
				apperr.WithCause(err),
			))
			return
		}
		credentialExpires = &parsed
	}

	err := h.syncSessionUseCase.Execute(c.Request.Context(), application.SyncSessionCommand{
		UserID:            req.UserID,
		LMSKey:            req.LMSKey,
		ExternalUserID:    req.ExternalUserID,
		APIDomain:         req.APIDomain,
		Metadata:          req.Metadata,
		CredentialSchema:  req.CredentialSchema,
		CredentialPayload: req.CredentialPayload,
		CredentialExpires: credentialExpires,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) isAuthorized(secret string) bool {
	if strings.TrimSpace(h.internalSyncSecret) == "" {
		return false
	}

	return secret == h.internalSyncSecret
}
