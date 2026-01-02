package auth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-ldap/ldap/v3"
)

type LDAPAuthenticator struct {
	config LDAPConfig
}

func NewLDAPAuthenticator(config LDAPConfig) (*LDAPAuthenticator, error) {
	if config.Server == "" {
		return nil, fmt.Errorf("LDAP server is required")
	}
	if config.BindDN == "" {
		return nil, fmt.Errorf("LDAP bind DN is required")
	}
	if config.UserBaseDN == "" {
		return nil, fmt.Errorf("LDAP user base DN is required")
	}
	if config.UserFilter == "" {
		config.UserFilter = "(uid=%s)"
	}
	if config.GroupFilter == "" {
		config.GroupFilter = "(member=%s)"
	}

	return &LDAPAuthenticator{config: config}, nil
}

func (l *LDAPAuthenticator) Authenticate(r *http.Request) error {
	username, password, ok := r.BasicAuth()
	if !ok {
		return fmt.Errorf("missing_credentials")
	}
	
	if username == "" || password == "" {
		return fmt.Errorf("invalid_credentials")
	}
	conn, err := ldap.DialURL(l.config.Server)
	if err != nil {
		log.Printf("Failed to connect to LDAP server: %v", err)
		return fmt.Errorf("authentication failed")
	}
	defer conn.Close()

	err = conn.Bind(l.config.BindDN, l.config.BindPassword)
	if err != nil {
		log.Printf("Failed to bind to LDAP server: %v", err)
		return fmt.Errorf("authentication failed")
	}

	searchRequest := ldap.NewSearchRequest(
		l.config.UserBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(l.config.UserFilter, ldap.EscapeFilter(username)),
		[]string{"dn"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		log.Printf("Failed to search for user %s: %v", username, err)
		return fmt.Errorf("authentication failed")
	}

	if len(sr.Entries) == 0 {
		log.Printf("User %s not found in LDAP", username)
		return fmt.Errorf("authentication failed")
	}

	if len(sr.Entries) > 1 {
		log.Printf("Multiple entries found for user %s", username)
		return fmt.Errorf("authentication failed")
	}

	userDN := sr.Entries[0].DN

	err = conn.Bind(userDN, password)
	if err != nil {
		log.Printf("Failed to authenticate user %s: %v", username, err)
		return fmt.Errorf("authentication failed")
	}

	if l.config.RequiredGroup != "" {
		err = conn.Bind(l.config.BindDN, l.config.BindPassword)
		if err != nil {
			log.Printf("Failed to rebind for group check: %v", err)
			return fmt.Errorf("authentication failed")
		}

		groupSearchRequest := ldap.NewSearchRequest(
			l.config.GroupBaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&%s(cn=%s))", fmt.Sprintf(l.config.GroupFilter, ldap.EscapeFilter(userDN)), ldap.EscapeFilter(l.getGroupCN(l.config.RequiredGroup))),
			[]string{"cn"},
			nil,
		)

		gsr, err := conn.Search(groupSearchRequest)
		if err != nil {
			log.Printf("Failed to search for group membership: %v", err)
			return fmt.Errorf("authentication failed")
		}

		if len(gsr.Entries) == 0 {
			log.Printf("User %s is not a member of required group %s", username, l.config.RequiredGroup)
			return fmt.Errorf("forbidden")
		}
	}

	return nil
}

func (l *LDAPAuthenticator) getGroupCN(groupDN string) string {
	entry, err := ldap.ParseDN(groupDN)
	if err != nil {
		return groupDN
	}
	
	for _, rdn := range entry.RDNs {
		for _, attr := range rdn.Attributes {
			if attr.Type == "cn" || attr.Type == "CN" {
				return attr.Value
			}
		}
	}
	
	return groupDN
}
