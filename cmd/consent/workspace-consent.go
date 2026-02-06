package consent

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/helper"
	"github.com/spf13/viper"
)

// CheckWorkspaceConsent checks if the user has consented to run commands
// from a workspace at the given directory path.
// Returns true if consent exists and is not expired.
func CheckWorkspaceConsent(workspaceDir string) bool {
	key := workspaceConsentKey(workspaceDir)
	secretValue, err := helper.GetSecret(key)
	if err != nil {
		return false
	}

	var consent Consent
	if err := json.Unmarshal([]byte(secretValue), &consent); err != nil {
		return false
	}

	if time.Unix(consent.ExpiresAt, 0).Before(time.Now()) {
		return false
	}

	return true
}

// RequestWorkspaceConsent prompts the user to trust a workspace.
// Displays the workspace path and asks for y/N confirmation.
// On approval, saves consent with expiration from USER_CONSENT_LIFE_KEY config.
func RequestWorkspaceConsent(workspaceDir string) bool {
	fmt.Printf("Workspace commands discovered at: %s\n", workspaceDir)
	console.Reminder("Do you trust and want to load commands from this workspace? [yN]")

	var resp int
	if _, err := fmt.Scanf("%c", &resp); err != nil || (resp != 'y' && resp != 'Y') {
		fmt.Printf("Workspace commands not loaded.\n")
		fmt.Printf("-----------------------------\n\n")
		return false
	}

	if err := SaveWorkspaceConsent(workspaceDir); err != nil {
		fmt.Printf("Warning: failed to save workspace consent: %v\n", err)
	}

	return true
}

// SaveWorkspaceConsent persists consent for a workspace directory.
func SaveWorkspaceConsent(workspaceDir string) error {
	keyLife := viper.GetDuration(config.USER_CONSENT_LIFE_KEY).Seconds()
	if keyLife <= 0 {
		// default: 30 days
		keyLife = 2592000
	}

	key := workspaceConsentKey(workspaceDir)
	secretValue, err := json.Marshal(Consent{
		ExpiresAt: time.Now().Unix() + int64(keyLife),
		Consents:  []string{"workspace"},
	})
	if err != nil {
		return err
	}

	return helper.SetSecret(key, string(secretValue))
}

// workspaceConsentKey returns a keychain key for the workspace directory.
// Uses sha256 hash of the absolute path to avoid special characters.
func workspaceConsentKey(workspaceDir string) string {
	hash := sha256.Sum256([]byte(workspaceDir))
	return fmt.Sprintf("workspace_consent_%x", hash[:8])
}
