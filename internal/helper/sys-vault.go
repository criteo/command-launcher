package helper

import (
	"runtime"

	vault "github.com/criteo/command-launcher/internal/gvault"
	log "github.com/sirupsen/logrus"
	"github.com/zalando/go-keyring"
)

const VAULT_SERVICE_NAME = "cdt"

func SetSecret(key string, value string) error {
	if runtime.GOOS == "linux" || HasDebugFlag(USE_FILE_VAULT) {
		return setSecretFromFileVault(key, value)
	}
	if err := keyring.Set(VAULT_SERVICE_NAME, key, value); err != nil {
		// fallback to the file vault
		log.Warnf("fail to write secret to system vault, fallback to file vault, %v\n", err)
		return setSecretFromFileVault(key, value)
	}
	return nil
}

func GetSecret(key string) (string, error) {
	if runtime.GOOS == "linux" || HasDebugFlag(USE_FILE_VAULT) {
		return getSecretFromFileVault(key)
	}

	secret, err := keyring.Get(VAULT_SERVICE_NAME, key)
	if err != nil {
		// fallback to the file vault
		log.Warnf("fail to get secret from system vault, fallback to file vault, %v\n", err)
		return getSecretFromFileVault(key)
	}
	return secret, nil
}

func setSecretFromFileVault(key string, value string) error {
	fv, err := vault.CreateVault(VAULT_SERVICE_NAME)
	if err != nil {
		return err
	}
	return fv.Write(key, value)
}

func getSecretFromFileVault(key string) (string, error) {
	fv, err := vault.CreateVault(VAULT_SERVICE_NAME)
	if err != nil {
		return "", err
	}
	return fv.Read(key)
}
