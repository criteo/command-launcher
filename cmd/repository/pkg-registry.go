package repository

import "github.com/criteo/command-launcher/internal/command"

type Registry interface {
	/* write interfaces */
	Load(repoDir string) error
	Add(pkg command.PackageManifest, repoDir string) error
	Remove(pkgName string, repoDir string) error
	Update(pkg command.PackageManifest, repoDir string) error

	/* read interfaces */
	AllPackages() []command.PackageManifest
	AllCommands() []command.Command
	GroupCommands() []command.Command
	ExecutableCommands() []command.Command
	SystemLoginCommand() command.Command
	SystemMetricsCommand() command.Command
	Package(name string) (command.PackageManifest, error)
	Command(group string, name string) (command.Command, error)
}
