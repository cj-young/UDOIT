package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"rewritetest/internal/lms/internal/domain"
	lmssqlc "rewritetest/internal/lms/internal/infrastructure/sqlc"
	"rewritetest/internal/shared/apperr"
)

type MySQLLMSObjectMappingRepository struct {
	queries *lmssqlc.Queries
}

func NewMySQLLMSObjectMappingRepository(db *sql.DB) *MySQLLMSObjectMappingRepository {
	return &MySQLLMSObjectMappingRepository{
		queries: lmssqlc.New(db),
	}
}

func (r *MySQLLMSObjectMappingRepository) GetByTypeAndInternalID(ctx context.Context, objectType domain.LMSObjectType, internalID int64) (*domain.LMSObjectMapping, error) {
	row, err := r.queries.GetLMSObjectMappingByTypeAndInternalID(ctx, lmssqlc.GetLMSObjectMappingByTypeAndInternalIDParams{
		ObjectType: string(objectType),
		InternalID: uint64(internalID),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	mappingData := map[string]any{}
	if len(row.MappingJson) > 0 {
		if err := json.Unmarshal(row.MappingJson, &mappingData); err != nil {
			return nil, err
		}
	}

	var externalID string
	if row.ExternalID.Valid {
		externalID = row.ExternalID.String
	} else {
		return nil, apperr.New(apperr.CodeInternal, "external_id_not_valid", "External ID is not valid")
	}

	mapping := domain.NewLMSObjectMapping(internalID, objectType, externalID, row.LmsKey, mappingData)
	return &mapping, nil
}

var _ domain.LMSObjectMappingRepository = (*MySQLLMSObjectMappingRepository)(nil)
