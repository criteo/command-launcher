package context

type LauncherContext interface {
	AppVersion() string

	AppName() string

	AppDirname() string

	UsernameVarEnv() string

	PasswordVarEnv() string

	DebugFlagsVarEnv() string

	ConfigurationFileVarEnv() string
}
