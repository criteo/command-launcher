package context

import (
	"fmt"
	"strings"
)

type defaultContext struct {
	appName string
}

var context = defaultContext{}

func InitContext(appName string) LauncherContext {
	// TODO check the appName value
	context.appName = appName
	return &context
}

func AppContext() (LauncherContext, error) {
	if context.appName != "" {
		return &context, nil
	}

	return nil, fmt.Errorf("uninitialized context, the root command has to init the context")
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

func (ctx *defaultContext) prefix() string {
	return strings.ToUpper(ctx.appName)
}
