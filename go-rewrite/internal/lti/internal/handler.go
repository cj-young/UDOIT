package internal

import (
	"context"
	"fmt"
	"net/http"

	"rewritetest/internal/auth"
	"rewritetest/internal/lti/internal/application"
	"rewritetest/internal/shared/apperr"

	"github.com/gin-gonic/gin"
)

type SessionCreator interface {
	CreateSession(ctx context.Context, userID int64, tenantID int64) (auth.Session, error)
}

type Handler struct {
	sessionCreator           SessionCreator
	getLaunchRedirectUseCase *application.GetLaunchRedirectUseCase
	processLaunchUseCase     *application.ProcessLaunchUseCase
	baseURL string
}

func NewHandler(sessionCreator SessionCreator, getLaunchRedirectUseCase *application.GetLaunchRedirectUseCase, processLaunchUseCase *application.ProcessLaunchUseCase, baseURL string) *Handler {
	return &Handler{
		sessionCreator:           sessionCreator,
		getLaunchRedirectUseCase: getLaunchRedirectUseCase,
		processLaunchUseCase:     processLaunchUseCase,
		baseURL: baseURL,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/authorize", h.handleLoginInitiation)
	rg.POST("/authorize/check", h.handleLaunch)
}

type LoginInitiationRequest struct {
	ISS           string `form:"iss" binding:"required"`
	LoginHint     string `form:"login_hint" binding:"required"`
	TargetLinkURI string `form:"target_link_uri" binding:"required"`
	ClientID      string `form:"client_id" binding:"required"`
	LTIMessageHint string `form:"lti_message_hint"`
}

func (h *Handler) handleLoginInitiation(c *gin.Context) {
	var req LoginInitiationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Error(apperr.New(
			apperr.CodeValidation, "invalid_request_payload", "Invalid request payload",
			apperr.WithOp("lti.handler.handleLoginInitiation"),
		))
		return
	}

	launchCallbackURL := fmt.Sprintf("%s/lti/authorize/check", h.baseURL)

	getLaunchRedirectQuery := application.GetLaunchRedirectQuery{
		Issuer:        req.ISS,
		LoginHint:     req.LoginHint,
		TargetLinkURI: req.TargetLinkURI,
		ClientID:      req.ClientID,
		RedirectURI:   launchCallbackURL,
		LTIMessageHint: req.LTIMessageHint,
	}

	redirectURL, err := h.getLaunchRedirectUseCase.Execute(c.Request.Context(), getLaunchRedirectQuery)
	if err != nil {
		if apperr.IsAppError(err) {
			c.Error(err)
			return
		}

		c.Error(apperr.New(
			apperr.CodeInternal, "launch_redirect_error", "Failed to get launch redirect URL",
			apperr.WithOp("lti.handler.handleLoginInitiation"),
			apperr.WithCause(err),
		))
		return
	}

	c.Redirect(302, redirectURL)
}

type LaunchCallbackRequest struct {
	IDToken string `form:"id_token" binding:"required"`
	State   string `form:"state" binding:"required"`
}

func (h *Handler) handleLaunch(c *gin.Context) {
	var req LaunchCallbackRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Error(apperr.New(
			apperr.CodeValidation, "invalid_request_payload", "Invalid request payload",
			apperr.WithOp("lti.handler.handleLaunch"),
		))
		return
	}

	processLaunchCommand := application.ProcessLaunchCommand{
		IDToken: req.IDToken,
		State:   req.State,
	}

	result, err := h.processLaunchUseCase.Execute(c.Request.Context(), processLaunchCommand)
	if err != nil {
		if apperr.IsAppError(err) {
			c.Error(err)
			return
		}

		c.Error(apperr.New(
			apperr.CodeInternal, "launch_processing_error", "Failed to process launch",
			apperr.WithOp("lti.handler.handleLaunch"),
			apperr.WithCause(err),
		))
		return
	}

	session, err := h.sessionCreator.CreateSession(c.Request.Context(), result.UserID, result.TenantID)
	if err != nil {
		if apperr.IsAppError(err) {
			c.Error(err)
			return
		}

		c.Error(apperr.New(
			apperr.CodeInternal, "session_creation_error", "Failed to create user session",
			apperr.WithOp("lti.handler.handleLaunch"),
			apperr.WithCause(err),
		))
		return
	}

	cookie := &http.Cookie{
		Name:     "AUTH_TOKEN",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  session.ExpiresAt,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(c.Writer, cookie)

	c.Redirect(302, result.RedirectURL)
}
