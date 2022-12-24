package backend

import (
	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/repository"
)

type Backend interface {
	// Find a command with its group name and command name.
	// For the root level executable command, the group is empty string.
	// For the group command, the name is empty
	//
	// The group and cmd could be an alias defined by the RenameCommand
	// function. In this case, the FindCommand function is able to return
	// the original command
	// It maps the (group, name) to (registry/repository, package, group, name)
	FindCommand(group string, name string) (command.Command, error)

	// Get all group commands
	GroupCommands() []command.Command

	// Get all executable commands
	ExecutableCommands() []command.Command

	// Rename a command with a new name
	RenameCommand(cmd command.Command, new_name string) error

	// Find a system command by its name
	FindSystemCommand(name string) (command.Command, error)

	// Return all repositories or an empty slice
	AllRepositories() []repository.PackageRepository

	// Return default local repository
	DefaultRepository() repository.PackageRepository

	// Return dropin local repsository
	DropinRepository() repository.PackageRepository
}