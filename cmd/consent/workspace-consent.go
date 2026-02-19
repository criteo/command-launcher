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
	consent, err := getWorkspaceConsent(workspaceDir)
	if err != nil {
		return false
	}
	return hasConsentValue(consent, "workspace")
}

// IsWorkspaceConsentDenied checks if the user has explicitly denied consent
// for a workspace. Returns true if a non-expired denial record exists.
func IsWorkspaceConsentDenied(workspaceDir string) bool {
	consent, err := getWorkspaceConsent(workspaceDir)
	if err != nil {
		return false
	}
	return hasConsentValue(consent, "denied")
}

// RequestWorkspaceConsent prompts the user to trust a workspace.
// Displays the workspace path and asks for y/N confirmation.
// On approval, saves consent with expiration from USER_CONSENT_LIFE_KEY config.
// On denial, saves denial with the same expiration.
func RequestWorkspaceConsent(workspaceDir string) bool {
	fmt.Printf("This command is provided by workspace: %s\n", workspaceDir)
	console.Reminder("Do you trust and want to run commands from this workspace? [yN]")

	var resp int
	if _, err := fmt.Scanf("%c", &resp); err != nil || (resp != 'y' && resp != 'Y') {
		fmt.Printf("Workspace command execution denied.\n")
		fmt.Printf("-----------------------------\n\n")
		if err := saveWorkspaceConsentRecord(workspaceDir, "denied"); err != nil {
			fmt.Printf("Warning: failed to save workspace denial: %v\n", err)
		}
		return false
	}

	if err := SaveWorkspaceConsent(workspaceDir); err != nil {
		fmt.Printf("Warning: failed to save workspace consent: %v\n", err)
	}

	return true
}

// SaveWorkspaceConsent persists consent for a workspace directory.
func SaveWorkspaceConsent(workspaceDir string) error {
	return saveWorkspaceConsentRecord(workspaceDir, "workspace")
}

func saveWorkspaceConsentRecord(workspaceDir string, consentType string) error {
	keyLife := viper.GetDuration(config.USER_CONSENT_LIFE_KEY).Seconds()
	if keyLife <= 0 {
		// default: 30 days
		keyLife = 2592000
	}

	key := workspaceConsentKey(workspaceDir)
	secretValue, err := json.Marshal(Consent{
		ExpiresAt: time.Now().Unix() + int64(keyLife),
		Consents:  []string{consentType},
	})
	if err != nil {
		return err
	}

	return helper.SetSecret(key, string(secretValue))
}

func getWorkspaceConsent(workspaceDir string) (*Consent, error) {
	key := workspaceConsentKey(workspaceDir)
	secretValue, err := helper.GetSecret(key)
	if err != nil {
		return nil, err
	}

	var consent Consent
	if err := json.Unmarshal([]byte(secretValue), &consent); err != nil {
		return nil, err
	}

	if time.Unix(consent.ExpiresAt, 0).Before(time.Now()) {
		return nil, fmt.Errorf("consent expired")
	}

	return &consent, nil
}

func hasConsentValue(consent *Consent, value string) bool {
	for _, c := range consent.Consents {
		if c == value {
			return true
		}
	}
	return false
}

// workspaceConsentKey returns a keychain key for the workspace directory.
// Uses sha256 hash of the absolute path to avoid special characters.
func workspaceConsentKey(workspaceDir string) string {
	hash := sha256.Sum256([]byte(workspaceDir))
	return fmt.Sprintf("workspace_consent_%x", hash[:8])
}
