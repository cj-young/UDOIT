package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"rewritetest/internal/shared/apperr"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestErrorHandler_MapsStatusAndPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "validation", err: apperr.New(apperr.CodeValidation, "invalid", "bad input"), wantStatus: http.StatusBadRequest, wantCode: "VALIDATION"},
		{name: "not found", err: apperr.New(apperr.CodeNotFound, "missing", "not found"), wantStatus: http.StatusNotFound, wantCode: "NOT_FOUND"},
		{name: "unauthorized", err: apperr.New(apperr.CodeUnauthorized, "unauth", "unauthorized"), wantStatus: http.StatusUnauthorized, wantCode: "UNAUTHORIZED"},
		{name: "fallback", err: errors.New("boom"), wantStatus: http.StatusInternalServerError, wantCode: "INTERNAL"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(ErrorHandler())
			r.GET("/", func(c *gin.Context) {
				c.Error(tc.err)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			r.ServeHTTP(rec, req)

			require.Equal(t, tc.wantStatus, rec.Code)
			require.Contains(t, rec.Body.String(), `"code":"`+tc.wantCode+`"`)
		})
	}
}
