package context

type LauncherContext interface {
	AppName() string

	AppDirname() string

	UsernameVarEnv() string

	PasswordVarEnv() string

	DebugFlagsVarEnv() string
}
