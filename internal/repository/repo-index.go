package repository

import "github.com/criteo/command-launcher/internal/command"

type RepoIndex interface {
	/* write interfaces */
	Load(repoDir string) error
	Add(pkg command.PackageManifest, repoDir string, pkgDirName string) error
	Remove(pkgName string, repoDir string) error
	Update(pkg command.PackageManifest, repoDir string, pkgDirName string) error

	/* read interfaces */
	AllPackages() []command.PackageManifest
	AllCommands() []command.Command
	GroupCommands() []command.Command
	ExecutableCommands() []command.Command
	SystemLoginCommand() command.Command
	SystemMetricsCommand() command.Command
	Package(name string) (command.PackageManifest, error)
	IsPackageLocked(name string) (bool, error)
	SetPackageLock(name string) error
	Command(pkg string, group string, name string) (command.Command, error)
}
