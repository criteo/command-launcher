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

func (ctx *defaultContext) UsernameVarEnv() string {
	return fmt.Sprintf("%s_%s", ctx.prefix(), "USERNAME")
}

func (ctx *defaultContext) PasswordVarEnv() string {
	return fmt.Sprintf("%s_%s", ctx.prefix(), "PASSWORD")
}

func (ctx *defaultContext) DebugFlagsVarEnv() string {
	return fmt.Sprintf("%s_%s", ctx.prefix(), "DEBUG_FLAGS")
}

func (ctx *defaultContext) ConfigurationFileVarEnv() string {
	return fmt.Sprintf("%s_%s", ctx.prefix(), "CONFIG_FILE")
}

func (ctx *defaultContext) prefix() string {
	return strings.ToUpper(ctx.appName)
}
