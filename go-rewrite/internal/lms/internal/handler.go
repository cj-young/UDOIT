package internal

import (
	"rewritetest/internal/lms/internal/application"
	"rewritetest/internal/shared/apperr"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	processOAuthRedirectUseCase *application.ProcessOAuthRedirectUseCase
}

func NewHandler(processOAuthRedirectUseCase *application.ProcessOAuthRedirectUseCase) *Handler {
	return &Handler{
		processOAuthRedirectUseCase: processOAuthRedirectUseCase,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/oauth/callback", h.handleOauthRedirect)
}

type OAuthRedirectRequest struct {
	Code  string `form:"code"`
	State string `form:"state"`
}

func (h *Handler) handleOauthRedirect(c *gin.Context) {
	var req OAuthRedirectRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Error(apperr.New(
			apperr.CodeValidation, "invalid_oauth_redirect_request", "The OAuth redirect request format is not valid",
			apperr.WithOp("lms.internal.handler.handleOauthRedirect"),
		))
		return
	}

	redirectURL, err := h.processOAuthRedirectUseCase.Execute(c.Request.Context(), req.State, req.Code)
	if err != nil {
		c.Error(err)
		return
	}

	c.Redirect(302, redirectURL)
}