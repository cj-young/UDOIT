package internal

import (
	"rewritetest/internal/files/internal/application"
	"rewritetest/internal/shared/apperr"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	getFileUseCase *application.GetFileUseCase
}

func NewHandler(getFileUseCase *application.GetFileUseCase) *Handler {
	return &Handler{
		getFileUseCase: getFileUseCase,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/:file/review", h.HandleReviewFile)
	rg.GET("/hello", h.HandleHelloFiles)
}

func (h *Handler) HandleHelloFiles(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello, Files!",
	})
}

func (h *Handler) HandleReviewFile(c *gin.Context) {
	fileID, err := strconv.ParseInt(c.Param("file"), 10, 64)
	if err != nil {
		c.Error(apperr.New(
			apperr.CodeValidation, "invalid_file_id", "File ID must be a valid integer",
			apperr.WithCause(err),
			apperr.WithOp("files.handler.HandleReviewFile"),
		))
	}

	_, err = h.getFileUseCase.Execute(c.Request.Context(), fileID)
	if err != nil {
		c.Error(apperr.New(
			apperr.CodeInternal, "failed_to_get_file", "Failed to retrieve file information",
			apperr.WithCause(err),
			apperr.WithOp("files.handler.HandleReviewFile"),
		))
		return
	}

}