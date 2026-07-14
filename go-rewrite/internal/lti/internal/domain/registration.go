package domain

type Registration struct {
	id                    int64
	Issuer                string
	ClientID              string
	DeploymentID          string
	LoginAuthEndpoint     string
	JWKEndpoint           string
	ServiceAuthEndpoint   string
	ServiceLogoutEndpoint string
}

func (r *Registration) ID() int64 {
	return r.id
}

func NewRegistration(issuer, clientID, deploymentID, loginAuthEndpoint, jwkEndpoint, serviceAuthEndpoint, serviceLogoutEndpoint string) *Registration {
	return &Registration{
		Issuer:                issuer,
		ClientID:              clientID,
		DeploymentID:          deploymentID,
		LoginAuthEndpoint:     loginAuthEndpoint,
		JWKEndpoint:           jwkEndpoint,
		ServiceAuthEndpoint:   serviceAuthEndpoint,
		ServiceLogoutEndpoint: serviceLogoutEndpoint,
	}
}

func RehydrateRegistration(id int64, issuer, clientID, deploymentID, loginAuthEndpoint, jwkEndpoint, serviceAuthEndpoint, serviceLogoutEndpoint string) *Registration {
	return &Registration{
		id:                    id,
		Issuer:                issuer,
		ClientID:              clientID,
		DeploymentID:          deploymentID,
		LoginAuthEndpoint:     loginAuthEndpoint,
		JWKEndpoint:           jwkEndpoint,
		ServiceAuthEndpoint:   serviceAuthEndpoint,
		ServiceLogoutEndpoint: serviceLogoutEndpoint,
	}
}
