package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"rewritetest/internal/lms/internal/domain"
)

type MySQLUserLMSMappingRepository struct {
	db *sql.DB
}

func NewMySQLUserLMSMappingRepository(db *sql.DB) *MySQLUserLMSMappingRepository {
	return &MySQLUserLMSMappingRepository{db: db}
}

func (r *MySQLUserLMSMappingRepository) Upsert(ctx context.Context, mapping domain.UserLMSMapping) error {
	metadataJSON, err := json.Marshal(mapping.Metadata())
	if err != nil {
		return err
	}

	query := `
		INSERT INTO lms_user_mapping (user_id, lms_key, external_user_id, api_domain, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			lms_key = VALUES(lms_key),
			external_user_id = VALUES(external_user_id),
			api_domain = VALUES(api_domain),
			metadata = VALUES(metadata),
			updated_at = NOW()
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		mapping.UserID(),
		mapping.LMSKey(),
		nullIfEmpty(mapping.ExternalUserID()),
		nullIfEmpty(mapping.APIDomain()),
		metadataJSON,
	)
	return err
}

func (r *MySQLUserLMSMappingRepository) GetByUserID(ctx context.Context, userID int64) (*domain.UserLMSMapping, error) {
	query := `
		SELECT user_id, lms_key, COALESCE(external_user_id, ''), COALESCE(api_domain, ''), COALESCE(metadata, '{}'), created_at, updated_at
		FROM lms_user_mapping
		WHERE user_id = ?
	`

	var (
		resultUserID   int64
		lmsKey         string
		externalUserID string
		apiDomain      string
		metadataRaw    []byte
		createdAt      time.Time
		updatedAt      time.Time
	)

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&resultUserID,
		&lmsKey,
		&externalUserID,
		&apiDomain,
		&metadataRaw,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	metadata := map[string]any{}
	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
			return nil, err
		}
	}

	mapping := domain.RehydrateUserLMSMapping(
		resultUserID,
		strings.TrimSpace(lmsKey),
		externalUserID,
		apiDomain,
		metadata,
		createdAt,
		updatedAt,
	)

	return &mapping, nil
}

func nullIfEmpty(value string) any {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil
	}
	return v
}
