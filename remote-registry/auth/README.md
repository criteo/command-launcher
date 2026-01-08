# Authentication Module

This module provides pluggable authentication for the remote registry.

## Supported Authentication Types

### LDAP Authentication

LDAP authentication validates users against an LDAP directory and optionally checks group membership.

#### Configuration Example

```yaml
auth:
  enabled: true
  type: ldap
  ldap:
    server: ldap://ldap.example.com:389
    bind_dn: cn=admin,dc=example,dc=com
    bind_password: secret
    user_base_dn: ou=users,dc=example,dc=com
    user_filter: "(uid=%s)"
    group_base_dn: ou=groups,dc=example,dc=com
    group_filter: "(member=%s)"
    required_group: cn=registry-users,ou=groups,dc=example,dc=com
```

#### Configuration Parameters

- `server`: LDAP server URL (e.g., `ldap://host:389` or `ldaps://host:636`)
- `bind_dn`: DN to use for binding to LDAP server
- `bind_password`: Password for bind DN
- `user_base_dn`: Base DN to search for users
- `user_filter`: LDAP filter to find user (use `%s` as username placeholder, defaults to `(uid=%s)`)
- `group_base_dn`: Base DN to search for groups (required if using `required_group`)
- `group_filter`: LDAP filter to check group membership (use `%s` as user DN placeholder, defaults to `(member=%s)`)
- `required_group`: DN of the group that users must be a member of (optional)

#### Authentication Flow

1. Client sends HTTP request with Basic Authentication header
2. Server connects to LDAP using configured bind credentials
3. Server searches for user using `user_filter` in `user_base_dn`
4. Server authenticates user by attempting to bind with user's DN and provided password
5. If `required_group` is set, server checks if user is a member of that group
6. Returns success (200) or failure (401/403)

#### HTTP Status Codes

- `401 Unauthorized`: Missing or invalid credentials
- `403 Forbidden`: Valid credentials but user not in required group

### Custom JWT Authentication

Custom JWT authentication delegates token validation to an external script. This allows integration with any JWT provider without modifying the registry code.

#### Configuration Example

```yaml
auth:
  enabled: true
  type: custom_jwt
  custom_jwt:
    script: /path/to/jwt-validator.sh
    required_group: my-required-group
```

#### Configuration Parameters

- `script`: Path to the external script that validates JWT tokens (required)
- `required_group`: Group name that must be present in the script output (optional)

#### Script Contract

The script receives the JWT token as its first argument and must follow this contract:

**On success (exit code 0):**
- Print one group name per line to stdout

**On failure (exit code 1):**
- Print error message to stderr

#### Example Script

```bash
#!/bin/bash
TOKEN="$1"

# Validate token and extract groups (example using jq)
PAYLOAD=$(echo "$TOKEN" | cut -d'.' -f2 | base64 -d 2>/dev/null)
if [ $? -ne 0 ]; then
    echo "Invalid token format" >&2
    exit 1
fi

# Extract and print groups
echo "$PAYLOAD" | jq -r '.groups[]' 2>/dev/null
if [ $? -ne 0 ]; then
    echo "Failed to extract groups" >&2
    exit 1
fi
```

#### Authentication Flow

1. Client sends HTTP request with `Authorization: Bearer <token>` header
2. Server extracts the token from the header
3. Server executes the configured script with the token as argument
4. Script validates the token and returns groups (exit 0) or error (exit 1)
5. If `required_group` is set, server checks if that group is in the script output
6. Returns success (200) or failure (401/403)

#### HTTP Status Codes

- `401 Unauthorized`: Missing or invalid Bearer token
- `403 Forbidden`: Valid token but user not in required group

### No Authentication

To disable authentication:

```yaml
auth:
  enabled: false
  type: none
```

Or simply omit the `auth` section from the configuration file.

## Protected Endpoints

All write operations require authentication when enabled:
- `POST /registry` - Create registry
- `PUT /registry/{registry}` - Update registry
- `DELETE /registry/{registry}` - Delete registry
- `POST /registry/{registry}/package` - Create package
- `PUT /registry/{registry}/package/{package}` - Update package
- `DELETE /registry/{registry}/package/{package}` - Delete package
- `POST /registry/{registry}/package/{package}/version` - Create package version
- `DELETE /registry/{registry}/package/{package}/version/{version}` - Delete package version

Read operations (GET) remain public.

## Usage Example

```bash
# Without authentication
curl http://localhost:8080/registry

# With authentication
curl -u username:password -X POST http://localhost:8080/registry \
  -H "Content-Type: application/json" \
  -d '{"name":"my-registry","description":"My Registry"}'
```

## Adding New Authentication Types

Each authenticator is responsible for extracting its own credentials from the HTTP request. The `Authenticator` interface requires implementing:

```go
type Authenticator interface {
    Authenticate(r *http.Request) error
}
```

To add a new authentication type:

1. Create a new authenticator struct that implements the `Authenticator` interface
2. Extract credentials from the request (e.g., Bearer token, API key header, etc.)
3. Return specific error types: `"missing_credentials"`, `"forbidden"`, or other error messages
4. Add configuration struct for the new type to `config.go`
5. Update the `NewAuthenticator` factory function in `auth.go` to handle the new type
6. Document the new type in this README

### Example: Token-Based Authentication

```go
type TokenAuthenticator struct {
    config TokenConfig
}

func (t *TokenAuthenticator) Authenticate(r *http.Request) error {
    token := r.Header.Get("Authorization")
    if token == "" {
        return fmt.Errorf("missing_credentials")
    }
    
    // Validate token
    if !t.validateToken(token) {
        return fmt.Errorf("invalid_credentials")
    }
    
    return nil
}
```
