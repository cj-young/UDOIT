package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"rewritetest/internal/lms/internal/domain"
)

type MySQLLMSObjectMappingRepository struct {
	db *sql.DB
}

func NewMySQLLMSObjectMappingRepository(db *sql.DB) *MySQLLMSObjectMappingRepository {
	return &MySQLLMSObjectMappingRepository{db: db}
}

func (r *MySQLLMSObjectMappingRepository) GetByTypeAndInternalID(ctx context.Context, objectType domain.LMSObjectType, internalID int64) (*domain.LMSObjectMapping, error) {
	query := `
		SELECT lms_key, mapping_json
		FROM lms_object_mapping
		WHERE object_type = ? AND internal_id = ?
	`

	var lmsKey string
	var mappingRaw []byte
	err := r.db.QueryRowContext(ctx, query, string(objectType), internalID).Scan(&lmsKey, &mappingRaw)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	mappingData := map[string]any{}
	if len(mappingRaw) > 0 {
		if err := json.Unmarshal(mappingRaw, &mappingData); err != nil {
			return nil, err
		}
	}

	mapping := domain.NewLMSObjectMapping(internalID, objectType, lmsKey, mappingData)
	return &mapping, nil
}

var _ domain.LMSObjectMappingRepository = (*MySQLLMSObjectMappingRepository)(nil)
