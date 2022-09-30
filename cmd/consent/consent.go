package consent

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/helper"
	"github.com/spf13/viper"
)

type Consent struct {
	ExpiresAt int64    `json:"expiresAt"`
	Consents  []string `json:"consents"`
}

const (
	USERNAME    = "USERNAME"
	PASSWORD    = "PASSWORD"
	LOG_LEVEL   = "LOG_LEVEL"
	LOGIN_TOKEN = "LOGIN_TOKEN"
	DEBUG_FLAGS = "DEBUG_FLAGS"
)

var AvailableConsents = []string{
	USERNAME, PASSWORD, LOGIN_TOKEN, LOG_LEVEL, DEBUG_FLAGS,
}

// GetConsents function returns the user consent of a particular command
// This function returns a list of agreed consents.
func GetConsents(cmdGroup string, cmdName string, requests []string, enabled bool) ([]string, error) {
	if !enabled {
		return AvailableConsents, nil
	}

	if len(requests) == 0 {
		return requests, nil
	}

	consent, err := getCmdConsents(cmdGroup, cmdName)
	if err == nil {
		return consent.Consents, nil
	}

	if requestConsent(cmdGroup, cmdName, requests) {
		// yes
		// expire in seconds for 30 days = 3600 * 24 * 30 = 2592000
		keyLife := viper.GetDuration(config.USER_CONSENT_LIFE_KEY).Seconds()
		if keyLife == 0 {
			keyLife = 2592000
		}
		if err := saveCmdConsents(cmdGroup, cmdName, requests, int64(keyLife)); err != nil {
			return requests, err
		}
		return requests, nil
	} // no

	return []string{}, nil
}

func getCmdConsents(cmdGroup string, cmdName string) (*Consent, error) {
	var err error
	secretKey := getConsentKey(cmdGroup, cmdName)
	secretValue := ""
	consent := Consent{}
	if secretValue, err = helper.GetSecret(secretKey); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(secretValue), &consent); err != nil {
		return nil, err
	}
	if time.Unix(consent.ExpiresAt, 0).Before(time.Now()) {
		// expired
		return nil, fmt.Errorf("consent expired")
	}
	return &consent, nil
}

func getConsentKey(cmdGroup string, cmdName string) string {
	return fmt.Sprintf("%s_%s", cmdGroup, cmdName)
}

func saveCmdConsents(cmdGroup string, cmdName string, requests []string, duration int64) error {
	secretKey := getConsentKey(cmdGroup, cmdName)
	secretValue, err := json.Marshal(Consent{
		// seconds for 30 days = 3600 * 24 * 30 = 2592000
		ExpiresAt: time.Now().Unix() + duration, //+ 2592000, // expires in 30 days
		Consents:  requests,
	})
	if err != nil {
		return err
	}
	if err := helper.SetSecret(secretKey, string(secretValue)); err != nil {
		return err
	}
	return nil
}

func requestConsent(cmdGroup string, cmdName string, requests []string) bool {
	fmt.Printf("Command '%s %s' requests access to the following resources:\n", cmdGroup, cmdName)
	for _, request := range requests {
		fmt.Printf("  - %s\n", request)
	}
	fmt.Println()
	console.Reminder("authorize the access? [yN]")
	var resp int
	if _, err := fmt.Scanf("%c", &resp); err != nil || (resp != 'y' && resp != 'Y') {
		fmt.Printf("Authorization refused by user\n")
		fmt.Printf("-----------------------------\n\n")
		return false
	}
	return true
}
