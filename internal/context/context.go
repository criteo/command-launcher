package context

type LauncherContext interface {
	AppVersion() string

	AppName() string

	AppDirname() string

	UsernameEnvVar() string

	PasswordEnvVar() string

	DebugFlagsEnvVar() string

	ConfigurationFileEnvVar() string

	RemoteConfigurationUrlEnvVar() string
}
