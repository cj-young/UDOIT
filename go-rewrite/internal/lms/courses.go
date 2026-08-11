package lms

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)


func (m *Module) GetCourseDataFromLTILaunch(ctx context.Context, tenantID int64, claims jwt.MapClaims, courseID int64) error {
	return nil
}