package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sharedmiddleware "rewritetest/internal/shared/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHandleSyncSession_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil, "expected-secret")
	r := gin.New()
	r.Use(sharedmiddleware.ErrorHandler())
	r.POST("/internal/lms/sync-session", handler.handleSyncSession)

	req := httptest.NewRequest(http.MethodPost, "/internal/lms/sync-session", nil)
	req.Header.Set("X-Internal-Sync-Secret", "wrong-secret")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
