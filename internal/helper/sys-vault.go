package helper

import (
	"runtime"

	"github.com/criteo/command-launcher/internal/context"
	vault "github.com/criteo/command-launcher/internal/gvault"
	log "github.com/sirupsen/logrus"
	"github.com/zalando/go-keyring"
)

func SetSecret(key string, value string) error {
	ctx, _ := context.AppContext()
	if runtime.GOOS == "linux" || HasDebugFlag(USE_FILE_VAULT) {
		return setSecretFromFileVault(key, value, ctx.AppName())
	}
	if err := keyring.Set(ctx.AppName(), key, value); err != nil {
		// fallback to the file vault
		log.Warnf("fail to write secret to system vault, fallback to file vault, %v\n", err)
		return setSecretFromFileVault(key, value, ctx.AppName())
	}
	return nil
}

func GetSecret(key string) (string, error) {
	ctx, _ := context.AppContext()
	if runtime.GOOS == "linux" || HasDebugFlag(USE_FILE_VAULT) {
		return getSecretFromFileVault(key, ctx.AppName())
	}

	secret, err := keyring.Get(ctx.AppName(), key)
	if err != nil {
		// fallback to the file vault
		log.Warnf("fail to get secret from system vault, fallback to file vault, %v\n", err)
		return getSecretFromFileVault(key, ctx.AppName())
	}
	return secret, nil
}

func setSecretFromFileVault(key string, value string, appName string) error {
	fv, err := vault.CreateVault(appName)
	if err != nil {
		return err
	}
	return fv.Write(key, value)
}

func getSecretFromFileVault(key string, appName string) (string, error) {
	fv, err := vault.CreateVault(appName)
	if err != nil {
		return "", err
	}
	return fv.Read(key)
}
