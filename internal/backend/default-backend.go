package backend

import (
	"errors"
	"fmt"
	"strings"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/repository"
)

// DefaultBackend supports multiple local repositories
type DefaultBackend struct {
	localRepoDirs []string

	localRepos []repository.PackageRepository
}

// Create a new default backend with multiple local repository directories
// When any of these repositories failed to load, an error is returned.
func NewDefaultBackend(localRepoDirs ...string) (*DefaultBackend, error) {
	backend := &DefaultBackend{
		localRepoDirs: localRepoDirs,
		localRepos:    []repository.PackageRepository{},
	}
	err := backend.load()
	return backend, err
}

func (backend *DefaultBackend) load() error {
	failures := []string{}
	for _, repoDir := range backend.localRepoDirs {
		repo, err := repository.CreateLocalRepository(repoDir, nil)
		if err != nil {
			failures = append(failures, err.Error())
		} else {
			backend.localRepos = append(backend.localRepos, repo)
		}
	}
	if len(failures) > 0 {
		return errors.New(fmt.Sprintf("failed to load repositories: %s", strings.Join(failures, "\n")))
	}
	return nil
}

/* Implement the Backend interface */

func (backend *DefaultBackend) FindCommand(group string, cmd string) (command.Command, error) {
	backend.localRepos[0].RepositoryFolder()
	return nil, nil
}

func (backend *DefaultBackend) RenameCommand(cmd command.Command, new_group string, new_cmd string) error {
	return nil
}

func (backend *DefaultBackend) FindSystemCommand(name string) (command.Command, error) {
	return nil, nil
}

func (backend *DefaultBackend) AllRepositories() []repository.PackageRepository {
	return backend.localRepos
}
