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

	Package(name string) (command.PackageManifest, error)

	Command(group string, name string) (command.Command, error)

	RepositoryFolder() (string, error)
}
