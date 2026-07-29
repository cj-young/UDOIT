package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"rewritetest/internal/lms/internal/domain"
	lmssqlc "rewritetest/internal/lms/internal/infrastructure/sqlc"
)

type MySQLLMSCredentialRepository struct {
	db *sql.DB
	queries *lmssqlc.Queries
}

func NewMySQLLMSCredentialRepository(db *sql.DB) *MySQLLMSCredentialRepository {
	return &MySQLLMSCredentialRepository{
		db: db,
		queries: lmssqlc.New(db),
	}
}

func (r *MySQLLMSCredentialRepository) UpsertActive(ctx context.Context, credential domain.LMSCredential) error {
	payloadJSON, err := json.Marshal(credential.Payload())
	if err != nil {
		return err
	}

	err = r.queries.UpsertLMSUserCredential(ctx, lmssqlc.UpsertLMSUserCredentialParams{
		UserID:        uint64(credential.UserID()),
		LmsKey:        string(credential.LMSKey()),
		SchemaName:   "",
		CredentialJson: payloadJSON,
		ExpiresAt:     nullableTime(credential.ExpiresAt()),
	})

	return err
}

func (r *MySQLLMSCredentialRepository) GetActiveByUser(ctx context.Context, userID int64) (*domain.LMSCredential, error) {
	query := `
		SELECT user_id, lms_key, credential_json, expires_at, is_active, created_at, updated_at
		FROM lms_user_credential
		WHERE user_id = ? AND is_active = 1
		LIMIT 1
	`

	var (
		resultUserID int64
		resultLMSKey string
		payloadRaw   []byte
		expiresAt    sql.NullTime
		isActive     bool
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&resultUserID,
		&resultLMSKey,
		&payloadRaw,
		&expiresAt,
		&isActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	payload := map[string]any{}
	if len(payloadRaw) > 0 {
		if err := json.Unmarshal(payloadRaw, &payload); err != nil {
			return nil, err
		}
	}

	credential, err := domain.RehydrateLMSCredential(
		resultUserID,
		resultLMSKey,
		payload,
		nullTimePtr(expiresAt),
		isActive,
		createdAt,
		updatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &credential, nil
}

func nullableTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{
		Time:  *value,
		Valid: true,
	}
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	copy := value.Time
	return &copy
}


var _ domain.LMSCredentialRepository = (*MySQLLMSCredentialRepository)(nil)
