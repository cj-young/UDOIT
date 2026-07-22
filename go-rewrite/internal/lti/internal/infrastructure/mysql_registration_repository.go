package infrastructure

import (
	"context"
	"database/sql"

	"rewritetest/internal/lti/internal/domain"
)

type MySQLRegistrationRepository struct {
	db *sql.DB
}

func NewMySQLRegistrationRepository(db *sql.DB) *MySQLRegistrationRepository {
	return &MySQLRegistrationRepository{
		db: db,
	}
}

func (r *MySQLRegistrationRepository) Create(ctx context.Context, registration domain.Registration) error {
	query := `
		INSERT INTO registration (issuer, client_id, tenant_id, login_auth_endpoint, jwk_endpoint, service_auth_endpoint, service_logout_endpoint)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, registration.Issuer, registration.ClientID, registration.TenantID, registration.LoginAuthEndpoint, registration.JWKEndpoint, registration.ServiceAuthEndpoint, registration.ServiceLogoutEndpoint)
	if err != nil {
		return err
	}
	return nil
}

func (r *MySQLRegistrationRepository) Save(ctx context.Context, registration domain.Registration) error {
	query := `
		UPDATE registration
		SET tenant_id = ?, login_auth_endpoint = ?, jwk_endpoint = ?, service_auth_endpoint = ?, service_logout_endpoint = ?
		WHERE issuer = ? AND client_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, registration.TenantID, registration.LoginAuthEndpoint, registration.JWKEndpoint, registration.ServiceAuthEndpoint, registration.ServiceLogoutEndpoint, registration.Issuer, registration.ClientID)
	if err != nil {
		return err
	}
	return nil
}

func (r *MySQLRegistrationRepository) GetByIssuerAndClientID(ctx context.Context, issuer string, clientID string) (*domain.Registration, error) {
	query := `
		SELECT issuer, client_id, tenant_id, login_auth_endpoint, jwk_endpoint, service_auth_endpoint, service_logout_endpoint
		FROM registration
		WHERE issuer = ? AND client_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, issuer, clientID)

	var registration domain.Registration
	err := row.Scan(&registration.Issuer, &registration.ClientID, &registration.TenantID, &registration.LoginAuthEndpoint, &registration.JWKEndpoint, &registration.ServiceAuthEndpoint, &registration.ServiceLogoutEndpoint)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &registration, nil
}
