package context

import (
	"fmt"
	"strings"
)

type defaultContext struct {
	appName    string
	appVersion string
	buildNum   string
}

var context = defaultContext{}

func InitContext(appName string, appVersion string, buildNum string) LauncherContext {
	// TODO check the appName value
	context.appName = appName
	context.appVersion = appVersion
	context.buildNum = buildNum

	return &context
}

func AppContext() (LauncherContext, error) {
	if context.appName != "" {
		return &context, nil
	}

	return nil, fmt.Errorf("uninitialized context, the root command has to init the context")
}

func (ctx *defaultContext) AppVersion() string {
	return ctx.appVersion
}

func (ctx *defaultContext) AppBuildNum() string {
	return ctx.buildNum
}

func (ctx *defaultContext) AppName() string {
	return ctx.appName
}

func (ctx *defaultContext) AppDirname() string {
	return fmt.Sprintf(".%s", ctx.appName)
}

func (ctx *defaultContext) AppHomeEnvVar() string {
	return ctx.envVarName("HOME")
}

func (ctx *defaultContext) UsernameEnvVar() string {
	return ctx.envVarName("USERNAME")
}

func (ctx *defaultContext) PasswordEnvVar() string {
	return ctx.envVarName("PASSWORD")
}

func (ctx *defaultContext) LogLevelEnvVar() string {
	return ctx.envVarName("LOG_LEVEL")
}

func (ctx *defaultContext) DebugFlagsEnvVar() string {
	return ctx.envVarName("DEBUG_FLAGS")
}

func (ctx *defaultContext) ConfigurationFileEnvVar() string {
	return ctx.envVarName("CONFIG_FILE")
}

func (ctx *defaultContext) RemoteConfigurationUrlEnvVar() string {
	return ctx.envVarName("REMOTE_CONFIG_URL")
}

func (ctx *defaultContext) envVarName(name string) string {
	return fmt.Sprintf("%s_%s", ctx.prefix(), name)
}

func (ctx *defaultContext) prefix() string {
	return strings.ToUpper(ctx.appName)
}
