package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"rewritetest/internal/auth/internal/domain"
	"rewritetest/internal/shared/apperr"
)

type MySQLSessionRepository struct {
	db *sql.DB
}

func NewMySQLSessionRepository(db *sql.DB) *MySQLSessionRepository {
	return &MySQLSessionRepository{
		db: db,
	}
}

func (r *MySQLSessionRepository) Create(ctx context.Context, session domain.Session) error {
	return apperr.New(
		apperr.CodeInternal,
		"function_not_implemented",
		"The MySQL session repository does not have an implementation for Create due to conflicts with the domain model and the old database.",
		apperr.WithOp("auth.infrastructure.mysql_session_repository.Create"),
	)
}

func (r *MySQLSessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	query := `
		SELECT uuid, data, 
		FROM user_session
		WHERE uuid = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var sessionID string
	var rawData sql.NullString
	err := row.Scan(&sessionID, &rawData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if !rawData.Valid || rawData.String == "" {
		return nil, nil
	}

	var payload struct {
		UserID *int64 `json:"userId"`
		// TenantID *int64 `json:"tenantId"`
	}

	if err := json.Unmarshal([]byte(rawData.String), &payload); err != nil {
		return nil, err
	}

	if payload.UserID == nil {
		return nil, nil
	}

	// TODO: 1234567890 is a placeholder for the tenant ID. Once launches are handled in the rewrite, this should be replaced.
	session := domain.NewSession(sessionID, *payload.UserID, 1234567890, time.Now(), time.Now().Add(1000*time.Hour*24))
	return &session, nil
}

func (r *MySQLSessionRepository) DeleteByID(ctx context.Context, id string) error {
	// Implement MySQL logic to delete a session by ID
	query := `
		DELETE FROM user_session
		WHERE uuid = ?
	`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}
