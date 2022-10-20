package repository

import "github.com/criteo/command-launcher/internal/command"

type Registry interface {
	Add(pkg command.PackageManifest) error

	Remove(name string) (command.PackageManifest, error)

	Update(pkg command.PackageManifest) (command.PackageManifest, error)

	Package(name string) (command.PackageManifest, error)

	AllPackages() []command.PackageManifest

	AllCommands() []command.Command

	GroupCommands() []command.Command

	ExecutableCommands() []command.Command
}
