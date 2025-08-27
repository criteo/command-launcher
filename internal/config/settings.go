package config

import (
	"fmt"
	"slices"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	LOG_ENABLED_KEY                      = "LOG_ENABLED"
	LOG_LEVEL_KEY                        = "LOG_LEVEL"
	SELF_UPDATE_ENABLED_KEY              = "SELF_UPDATE_ENABLED"
	SELF_UPDATE_TIMEOUT_KEY              = "SELF_UPDATE_TIMEOUT"
	SELF_UPDATE_LATEST_VERSION_URL_KEY   = "SELF_UPDATE_LATEST_VERSION_URL"
	SELF_UPDATE_BASE_URL_KEY             = "SELF_UPDATE_BASE_URL"
	SELF_UPDATE_POLICY_KEY               = "SELF_UPDATE_POLICY"
	COMMAND_UPDATE_ENABLED_KEY           = "COMMAND_UPDATE_ENABLED"
	COMMAND_REPOSITORY_BASE_URL_KEY      = "COMMAND_REPOSITORY_BASE_URL"
	LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY = "LOCAL_COMMAND_REPOSITORY_DIRNAME"
	USAGE_METRICS_ENABLED_KEY            = "USAGE_METRICS_ENABLED"
	METRIC_GRAPHITE_HOST_KEY             = "METRIC_GRAPHITE_HOST"
	DEBUG_FLAGS_KEY                      = "DEBUG_FLAGS"
	DROPIN_FOLDER_KEY                    = "DROPIN_FOLDER"
	CI_ENABLED_KEY                       = "CI_ENABLED"
	PACKAGE_LOCK_FILE_KEY                = "PACKAGE_LOCK_FILE"
	ENABLE_USER_CONSENT_KEY              = "ENABLE_USER_CONSENT"
	USER_CONSENT_LIFE_KEY                = "USER_CONSENT_LIFE"
	SYSTEM_PACKAGE_KEY                   = "SYSTEM_PACKAGE"                 // the system package name
	SYSTEM_PACKAGE_PUBLIC_KEY_KEY        = "SYSTEM_PACKAGE_PUBLIC_KEY"      // the public key to verify system package
	SYSTEM_PACKAGE_PUBLIC_KEY_FILE_KEY   = "SYSTEM_PACKAGE_PUBLIC_KEY_FILE" // the public key file to verify system package
	VERIFY_PACKAGE_CHECKSUM_KEY          = "VERIFY_PACKAGE_CHECKSUM"
	VERIFY_PACKAGE_SIGNATURE_KEY         = "VERIFY_PACKAGE_SIGNATURE"
	EXTRA_REMOTES_KEY                    = "EXTRA_REMOTES"
	EXTRA_REMOTE_BASE_URL_KEY            = "REMOTE_BASE_URL"
	EXTRA_REMOTE_REPOSITORY_DIR_KEY      = "REPOSITORY_DIR"
	EXTRA_REMOTE_SYNC_POLICY_KEY         = "SYNC_POLICY"
	ENABLE_PACKAGE_SETUP_HOOK_KEY        = "ENABLE_PACKAGE_SETUP_HOOK"
	GROUP_HELP_BY_REGISTRY_KEY           = "GROUP_HELP_BY_REGISTRY"

	// internal commands are the commands with start partition number > INTERNAL_START_PARTITION
	INTERNAL_COMMAND_ENABLED_KEY = "INTERNAL_COMMAND_ENABLED"
	// experimental commands are the commands with start partition number > EXPERIMENTAL_START_PARTITION
	EXPERIMENTAL_COMMAND_ENABLED_KEY = "EXPERIMENTAL_COMMAND_ENABLED"
)

type ExtraRemote struct {
	Name          string `mapstructure:"name" json:"name"`
	RemoteBaseUrl string `mapstructure:"remote_base_url" json:"remote_base_url"`
	RepositoryDir string `mapstructure:"repository_dir" json:"repository_dir"`
	SyncPolicy    string `mapstructure:"sync_policy" json:"sync_policy"`
}

// SelfUpdatePolicy defines the behavior for self-update version comparison
type SelfUpdatePolicy string

const (
	SelfUpdatePolicyExactMatch SelfUpdatePolicy = "exact_match"        // Update if versions are different (legacy behavior)
	SelfUpdatePolicyOnlyNewer  SelfUpdatePolicy = "only_newer_version" // Update if remote is newer
)

var SettingKeys []string

func init() {
	SettingKeys = append([]string{},
		LOG_ENABLED_KEY,
		LOG_LEVEL_KEY,
		SELF_UPDATE_ENABLED_KEY,
		SELF_UPDATE_TIMEOUT_KEY,
		SELF_UPDATE_LATEST_VERSION_URL_KEY,
		SELF_UPDATE_BASE_URL_KEY,
		SELF_UPDATE_POLICY_KEY,
		COMMAND_UPDATE_ENABLED_KEY,
		COMMAND_REPOSITORY_BASE_URL_KEY,
		LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY,
		USAGE_METRICS_ENABLED_KEY,
		METRIC_GRAPHITE_HOST_KEY,
		DEBUG_FLAGS_KEY,
		DROPIN_FOLDER_KEY,
		CI_ENABLED_KEY,
		PACKAGE_LOCK_FILE_KEY,
		INTERNAL_COMMAND_ENABLED_KEY,
		EXPERIMENTAL_COMMAND_ENABLED_KEY,
		ENABLE_USER_CONSENT_KEY,
		USER_CONSENT_LIFE_KEY,
		SYSTEM_PACKAGE_KEY,
		SYSTEM_PACKAGE_PUBLIC_KEY_FILE_KEY,
		ENABLE_PACKAGE_SETUP_HOOK_KEY,
		GROUP_HELP_BY_REGISTRY_KEY,
	)
}

func SetSettingValue(key string, value string) error {
	upperKey := strings.ToUpper(key)
	switch upperKey {
	case LOG_ENABLED_KEY:
		return setBooleanConfig(upperKey, value)
	case LOG_LEVEL_KEY:
		return setLogLevelConfig(value)
	case SELF_UPDATE_ENABLED_KEY:
		return setBooleanConfig(upperKey, value)
	case SELF_UPDATE_TIMEOUT_KEY:
		return setDurationConfig(upperKey, value)
	case SELF_UPDATE_BASE_URL_KEY:
		return setStringConfig(upperKey, value)
	case SELF_UPDATE_LATEST_VERSION_URL_KEY:
		return setStringConfig(upperKey, value)
	case SELF_UPDATE_POLICY_KEY:
		return setSelfUpdatePolicyConfig(value)
	case COMMAND_UPDATE_ENABLED_KEY:
		return setBooleanConfig(upperKey, value)
	case COMMAND_REPOSITORY_BASE_URL_KEY:
		return setStringConfig(upperKey, value)
	case LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY:
		return setStringConfig(upperKey, value)
	case USAGE_METRICS_ENABLED_KEY:
		return setBooleanConfig(upperKey, value)
	case METRIC_GRAPHITE_HOST_KEY:
		return setStringConfig(upperKey, value)
	case DEBUG_FLAGS_KEY:
		return setStringConfig(upperKey, value)
	case DROPIN_FOLDER_KEY:
		return setStringConfig(upperKey, value)
	case CI_ENABLED_KEY:
		return setBooleanConfig(upperKey, value)
	case PACKAGE_LOCK_FILE_KEY:
		return setStringConfig(upperKey, value)
	case EXPERIMENTAL_COMMAND_ENABLED_KEY:
		return setBooleanConfig(upperKey, value)
	case INTERNAL_COMMAND_ENABLED_KEY:
		return setBooleanConfig(upperKey, value)
	case ENABLE_USER_CONSENT_KEY:
		return setBooleanConfig(upperKey, value)
	case USER_CONSENT_LIFE_KEY:
		return setDurationConfig(upperKey, value)
	case SYSTEM_PACKAGE_KEY:
		return setStringConfig(upperKey, value)
	case SYSTEM_PACKAGE_PUBLIC_KEY_KEY:
		return setStringConfig(upperKey, value)
	case SYSTEM_PACKAGE_PUBLIC_KEY_FILE_KEY:
		return setStringConfig(upperKey, value)
	case VERIFY_PACKAGE_CHECKSUM_KEY:
		return setBooleanConfig(upperKey, value)
	case VERIFY_PACKAGE_SIGNATURE_KEY:
		return setBooleanConfig(upperKey, value)
	case ENABLE_PACKAGE_SETUP_HOOK_KEY:
		return setBooleanConfig(upperKey, value)
	case GROUP_HELP_BY_REGISTRY_KEY:
		return setBooleanConfig(upperKey, value)
	}

	return fmt.Errorf("unsupported config %s", key)
}

func AddRemote(name, repoDir, remoteBaseUrl string, syncPolicy string) error {
	remotes := []ExtraRemote{}
	err := viper.UnmarshalKey(EXTRA_REMOTES_KEY, &remotes)
	if err != nil {
		return err
	}

	if syncPolicy != "never" && syncPolicy != "hourly" &&
		syncPolicy != "daily" && syncPolicy != "weekly" &&
		syncPolicy != "monthly" && syncPolicy != "always" {
		syncPolicy = "always"
	}

	for _, remote := range remotes {
		if remote.RemoteBaseUrl == remoteBaseUrl || remote.Name == name {
			return fmt.Errorf("remote already exists")
		}
	}

	remotes = append(remotes, ExtraRemote{
		Name:          name,
		RemoteBaseUrl: remoteBaseUrl,
		RepositoryDir: repoDir,
		SyncPolicy:    syncPolicy,
	})
	viper.Set(EXTRA_REMOTES_KEY, remotes)

	return nil
}

func RemoveRemote(name string) error {
	remotes := []ExtraRemote{}
	err := viper.UnmarshalKey(EXTRA_REMOTES_KEY, &remotes)
	if err != nil {
		return err
	}

	new_remotes := []ExtraRemote{}
	for _, remote := range remotes {
		if remote.Name != name {
			new_remotes = append(new_remotes, remote)
		}
	}

	viper.Set(EXTRA_REMOTES_KEY, new_remotes)
	return nil
}

func Remotes() ([]ExtraRemote, error) {
	remotes := []ExtraRemote{}
	err := viper.UnmarshalKey(EXTRA_REMOTES_KEY, &remotes)
	if err != nil {
		return remotes, err
	}
	return remotes, nil
}

func setBooleanConfig(key string, value string) error {
	if value == "true" {
		viper.Set(key, true)
		return nil
	} else if value == "false" {
		viper.Set(key, false)
		return nil
	}
	return fmt.Errorf("invalid format for boolean type")
}

func setDurationConfig(key string, value string) error {
	d, _ := time.ParseDuration(value)
	viper.Set(key, d)
	return nil
}

func setStringConfig(key string, value string) error {
	viper.Set(key, value)
	return nil
}

func setLogLevelConfig(value string) error {
	_, err := log.ParseLevel(strings.ToLower(value))
	if err == nil {
		viper.Set(LOG_LEVEL_KEY, strings.ToLower(value))
	}
	return err
}

func setChoicesConfig(key string, value string, label string, choices []string) error {
	if slices.Contains(choices, value) {
		setStringConfig(key, value)
		return nil
	}
	return fmt.Errorf("invalid value for %s: \"%s\"", label, value)
}

func setSelfUpdatePolicyConfig(value string) error {
	choices := []string{
		string(SelfUpdatePolicyExactMatch),
		string(SelfUpdatePolicyOnlyNewer),
	}
	return setChoicesConfig(SELF_UPDATE_POLICY_KEY, value, "self-update policy", choices)
}
