package repository

import "github.com/criteo/command-launcher/internal/command"

/*
PackageRepository is responsible for managing the local installed packages.

You can get the installed command from a package repository.

Note: package repository manages packages, a packages contains multiple commands.
*/
type PackageRepository interface {
	Install(pkg command.Package) error

	Uninstall(name string) error

	Update(pkg command.Package) error

	InstalledPackages() []command.PackageManifest

	InstalledCommands() []command.Command

	InstalledGroupCommands() []command.Command

	InstalledExecutableCommands() []command.Command

	InstalledSystemCommands() SystemCommands

	Package(name string) (command.PackageManifest, error)

	Command(group string, name string) (command.Command, error)

	RepositoryFolder() (string, error)
}

/*
System commands
*/
const (
	SYSTEM_LOGIN_COMMAND   = "__login__"
	SYSTEM_METRICS_COMMAND = "__metrics__"
)

type SystemCommands struct {
	/*
		login hook to extend the login process
		it is called during built-in "login" command execution
		login command accepts two arguments
		- arg1: username
		- arg2: password
		and it returns a json, contains all credentials, ex:
		{
			"username": "",
			"password": "",
			"login_token": ""
		}
		To reference these credentials, use environment variable:
		COLA_LOGIN_[CREDENTIAL_NAME]
	*/
	Login command.Command
	// send metrics hook
	// it is called at the the end of the command execution
	// the metrics command must provide following subcommands:
	// - metrics send
	Metrics command.Command
}
