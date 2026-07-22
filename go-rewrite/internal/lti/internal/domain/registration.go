package domain

type Registration struct {
	id                    int64
	Issuer                string
	ClientID              string
	LoginAuthEndpoint     string
	JWKEndpoint           string
	ServiceAuthEndpoint   string
	ServiceLogoutEndpoint string
	TenantID int64
}

func (r *Registration) ID() int64 {
	return r.id
}

func NewRegistration(issuer, clientID, loginAuthEndpoint, jwkEndpoint, serviceAuthEndpoint, serviceLogoutEndpoint string, tenantID int64) *Registration {
	return &Registration{
		Issuer:                issuer,
		ClientID:              clientID,
		LoginAuthEndpoint:     loginAuthEndpoint,
		JWKEndpoint:           jwkEndpoint,
		ServiceAuthEndpoint:   serviceAuthEndpoint,
		ServiceLogoutEndpoint: serviceLogoutEndpoint,
		TenantID:              tenantID,
	}
}

func RehydrateRegistration(id int64, issuer, clientID, loginAuthEndpoint, jwkEndpoint, serviceAuthEndpoint, serviceLogoutEndpoint string, tenantID int64) *Registration {
	return &Registration{
		id:                    id,
		Issuer:                issuer,
		ClientID:              clientID,
		LoginAuthEndpoint:     loginAuthEndpoint,
		JWKEndpoint:           jwkEndpoint,
		ServiceAuthEndpoint:   serviceAuthEndpoint,
		ServiceLogoutEndpoint: serviceLogoutEndpoint,
		TenantID:              tenantID,
	}
}
