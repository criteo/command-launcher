package backend

import (
	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/repository"
)

var RESERVED_CMD_SEARCH_KEY map[string]bool = map[string]bool{
	"#login":      true,
	"#package":    true,
	"#remote":     true,
	"#update":     true,
	"#rename":     true,
	"#config":     true,
	"#version":    true,
	"#help":       true,
	"#completion": true,
}

type Backend interface {
	// Load all managed repositories
	Reload() error
	// Find a command with its group name and command name.
	// For the root level executable command, the group is empty string.
	// For the group command, the name is empty
	//
	// The group and cmd could be an alias defined by the RenameCommand
	// function. In this case, the FindCommand function is able to return
	// the original command
	// It maps the (group, name) to (registry/repository, package, group, name)
	FindCommand(group string, name string) (command.Command, error)

	FindCommandByFullName(fullName string) (command.Command, error)

	// Get all group commands
	GroupCommands() []command.Command

	// Get all executable commands
	ExecutableCommands() []command.Command

	// Get system command by name
	SystemCommand(name string) command.Command

	// Rename a command with a new name
	RenameCommand(cmd command.Command, new_name string) error

	// Get all renamed commands
	AllRenamedCommands() map[string]string

	// Find a system command by its name
	FindSystemCommand(name string) (command.Command, error)

	// Get all packages sources managed by this backend
	AllPackageSources() []*PackageSource

	// Return all repositories or an empty slice
	AllRepositories() []repository.PackageRepository

	// Return default local repository
	DefaultRepository() repository.PackageRepository

	// Return dropin local repsository
	DropinRepository() repository.PackageRepository

	// start the server as daemon
	Serve(port int) error

	// Print out the command resolution details
	Debug()
}
