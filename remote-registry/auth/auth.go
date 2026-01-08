package auth

import (
	"fmt"
	"net/http"
)

type Authenticator interface {
	Authenticate(r *http.Request) error
}

func NewAuthenticator(config Config) (Authenticator, error) {
	if !config.Enabled {
		return nil, nil
	}

	switch config.Type {
	case "ldap":
		return NewLDAPAuthenticator(config.LDAP)
	case "custom_jwt":
		return NewCustomJWTAuthenticator(config.CustomJWT)
	case "none":
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown auth type: %s (valid options: ldap, custom_jwt, none)", config.Type)
	}
}
