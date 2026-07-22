package application

import (
	"context"
	"strings"

	"rewritetest/internal/lti/internal/domain"
	"rewritetest/internal/shared/apperr"
)

type RegisterRegistrationUseCase struct {
	registrationRepository domain.RegistrationRepository
}

type RegisterRegistrationCommand struct {
	Issuer                string
	ClientID              string
	TenantID              int64
	LoginAuthEndpoint     string
	JWKEndpoint           string
	ServiceAuthEndpoint   string
	ServiceLoginEndpoint string
}

func NewRegisterRegistrationUseCase(registrationRepository domain.RegistrationRepository) *RegisterRegistrationUseCase {
	return &RegisterRegistrationUseCase{registrationRepository: registrationRepository}
}

func (u *RegisterRegistrationUseCase) Execute(ctx context.Context, cmd RegisterRegistrationCommand) error {
	registration, err := toRegistration(cmd)
	if err != nil {
		return err
	}

	existing, err := u.registrationRepository.GetByIssuerAndClientID(ctx, registration.Issuer, registration.ClientID)
	if err != nil {
		return err
	}

	if existing == nil {
		return u.registrationRepository.Create(ctx, *registration)
	}

	return u.registrationRepository.Save(ctx, *registration)
}

func toRegistration(cmd RegisterRegistrationCommand) (*domain.Registration, error) {
	issuer := strings.TrimSpace(cmd.Issuer)
	clientID := strings.TrimSpace(cmd.ClientID)
	loginAuthEndpoint := strings.TrimSpace(cmd.LoginAuthEndpoint)
	jwkEndpoint := strings.TrimSpace(cmd.JWKEndpoint)
	serviceAuthEndpoint := strings.TrimSpace(cmd.ServiceAuthEndpoint)
	serviceLoginEndpoint := strings.TrimSpace(cmd.ServiceLoginEndpoint)

	if issuer == "" {
		return nil, apperr.New(
			apperr.CodeValidation,
			"invalid_issuer",
			"Issuer is required",
			apperr.WithOp("lti.application.register_registration.toRegistration"),
		)
	}

	if clientID == "" {
		return nil, apperr.New(
			apperr.CodeValidation,
			"invalid_client_id",
			"Client ID is required",
			apperr.WithOp("lti.application.register_registration.toRegistration"),
		)
	}

	if cmd.TenantID <= 0 {
		return nil, apperr.New(
			apperr.CodeValidation,
			"invalid_tenant_id",
			"Tenant ID must be greater than zero",
			apperr.WithOp("lti.application.register_registration.toRegistration"),
		)
	}

	if loginAuthEndpoint == "" || jwkEndpoint == "" || serviceAuthEndpoint == "" || serviceLoginEndpoint == "" {
		return nil, apperr.New(
			apperr.CodeValidation,
			"invalid_registration_endpoints",
			"All registration endpoint URLs are required",
			apperr.WithOp("lti.application.register_registration.toRegistration"),
		)
	}

	return domain.NewRegistration(
		issuer,
		clientID,
		loginAuthEndpoint,
		jwkEndpoint,
		serviceAuthEndpoint,
		serviceLoginEndpoint,
		cmd.TenantID,
	), nil
}
