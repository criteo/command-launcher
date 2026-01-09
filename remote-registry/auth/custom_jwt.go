package auth

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type CustomJWTAuthenticator struct {
	config CustomJWTConfig
}

func NewCustomJWTAuthenticator(config CustomJWTConfig) (*CustomJWTAuthenticator, error) {
	if config.Script == "" {
		return nil, fmt.Errorf("custom_jwt script is required")
	}
	if _, err := exec.LookPath(config.Script); err != nil {
		return nil, fmt.Errorf("custom_jwt script not found or not executable: %v", err)
	}

	return &CustomJWTAuthenticator{config: config}, nil
}

func (c *CustomJWTAuthenticator) Authenticate(r *http.Request) error {
	token, err := c.extractBearerToken(r)
	if err != nil {
		return err
	}

	groups, err := c.executeScript(token)
	if err != nil {
		return err
	}

	if c.config.RequiredGroup != "" {
		if !c.hasGroup(groups, c.config.RequiredGroup) {
			log.Printf("User is not a member of required group %s", c.config.RequiredGroup)
			return fmt.Errorf("forbidden")
		}
	}

	return nil
}

func (c *CustomJWTAuthenticator) extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing_credentials")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("invalid_credentials")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("invalid_credentials")
	}

	return token, nil
}

func (c *CustomJWTAuthenticator) executeScript(token string) ([]string, error) {
	cmd := exec.Command(c.config.Script, token)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Printf("Script failed with exit code %d: %s", exitError.ExitCode(), strings.TrimSpace(stderr.String()))
			return nil, fmt.Errorf("authentication failed")
		}
		log.Printf("Failed to execute script: %v", err)
		return nil, fmt.Errorf("authentication failed")
	}

	groups := c.parseGroups(stdout.String())
	return groups, nil
}

func (c *CustomJWTAuthenticator) parseGroups(output string) []string {
	var groups []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		group := strings.TrimSpace(line)
		if group != "" {
			groups = append(groups, group)
		}
	}
	return groups
}

func (c *CustomJWTAuthenticator) hasGroup(groups []string, requiredGroup string) bool {
	for _, group := range groups {
		if group == requiredGroup {
			return true
		}
	}
	return false
}
