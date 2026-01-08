package auth

type Config struct {
	Enabled   bool            `mapstructure:"enabled"`
	Type      string          `mapstructure:"type"`
	LDAP      LDAPConfig      `mapstructure:"ldap"`
	CustomJWT CustomJWTConfig `mapstructure:"custom_jwt"`
}

type LDAPConfig struct {
	Server        string `mapstructure:"server"`
	BindDN        string `mapstructure:"bind_dn"`
	BindPassword  string `mapstructure:"bind_password"`
	UserBaseDN    string `mapstructure:"user_base_dn"`
	UserFilter    string `mapstructure:"user_filter"`
	GroupBaseDN   string `mapstructure:"group_base_dn"`
	GroupFilter   string `mapstructure:"group_filter"`
	RequiredGroup string `mapstructure:"required_group"`
}

type CustomJWTConfig struct {
	Script        string `mapstructure:"script"`
	RequiredGroup string `mapstructure:"required_group"`
}
