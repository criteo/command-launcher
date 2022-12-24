package backend

import (
	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/repository"
)

type Backend interface {
	// Find a command with its group name and command name.
	// For root level command, the group is empty string.
	//
	// The group and cmd could be an alias defined by the RenameCommand
	// function. In this case, the FindCommand function is able to return
	// the original command
	FindCommand(group string, cmd string) (command.Command, error)

	// Rename a command with a new group and command name
	RenameCommand(cmd command.Command, new_group string, new_cmd string) error

	// Find a system command by its name
	FindSystemCommand(name string) (command.Command, error)

	// Return all repositories or an empty slice
	AllRepositories() []repository.PackageRepository
}
