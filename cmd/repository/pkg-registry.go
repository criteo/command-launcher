package repository

import "github.com/criteo/command-launcher/internal/command"

type Registry interface {
	Add(pkg command.PackageManifest) error

	Remove(pkgName string) error

	Update(pkg command.PackageManifest) error

	AllPackages() []command.PackageManifest

	AllCommands() []command.Command

	GroupCommands() []command.Command

	ExecutableCommands() []command.Command

	Package(name string) (command.PackageManifest, error)

	Command(group string, name string) (command.Command, error)
}
