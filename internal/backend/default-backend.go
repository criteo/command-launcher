package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/repository"
	"gopkg.in/yaml.v3"
)

// DefaultBackend supports multiple local repositories
// It contains:
// - 1 dropin local repository - index 0
// - 1 default managed repository - index 1
// - n additional managed repository - index 2 ..
type DefaultBackend struct {
	homeDir  string
	repoDirs []string
	repos    []repository.PackageRepository

	cmdsCache      map[string]command.Command
	groupCmds      map[string]command.Command
	executableCmds map[string]command.Command

	alias map[string]string
}

const DROPIN_REPO_INDEX = 0
const DEFAULT_REPO_INDEX = 1
const DEFAULT_REPO_ID = "default"
const DROPIN_REPO_ID = "dropin"

// Create a new default backend with multiple local repository directories
// When any of these repositories failed to load, an error is returned.
func NewDefaultBackend(appHomeDir string, dropinRepoDir string, defaultRepoDir string, additionalRepoDirs ...string) (*DefaultBackend, error) {
	backend := &DefaultBackend{
		homeDir:   appHomeDir,
		repoDirs:  append([]string{dropinRepoDir, defaultRepoDir}, additionalRepoDirs...),
		repos:     []repository.PackageRepository{},
		cmdsCache: map[string]command.Command{},
		alias:     map[string]string{},
	}
	err := backend.load()
	return backend, err
}

func (backend *DefaultBackend) load() error {
	err := backend.loadRepos()
	backend.loadAlias()
	backend.extractCmds()
	return err
}

func (backend *DefaultBackend) loadRepos() error {
	failures := []string{}
	for i, repoDir := range backend.repoDirs {
		repoID := ""
		switch i {
		case DROPIN_REPO_INDEX:
			repoID = DROPIN_REPO_ID
		case DEFAULT_REPO_INDEX:
			repoID = DEFAULT_REPO_ID
		default:
			repoID = fmt.Sprintf("repo%d", i-1)
		}
		repo, err := repository.CreateLocalRepository(repoID, repoDir, nil)
		if err != nil {
			failures = append(failures, err.Error())
		} else {
			backend.repos = append(backend.repos, repo)
		}
	}
	if len(failures) > 0 {
		return errors.New(fmt.Sprintf("failed to load repositories: %s", strings.Join(failures, "\n")))
	}
	return nil
}

func (backend *DefaultBackend) loadAlias() error {
	if aliasFile, err := os.Open(filepath.Join(backend.homeDir, "alias.json")); err == nil {
		defer aliasFile.Close()
		if err != nil {
			return fmt.Errorf("no such alias file found (%s)", err)
		}

		stat, err := aliasFile.Stat()
		if err != nil {
			return fmt.Errorf("cannot read alias file (%s)", err)
		}

		var payload = make([]byte, stat.Size())
		nb, err := aliasFile.Read(payload)
		if err != nil && err != io.EOF || nb != int(stat.Size()) {
			return fmt.Errorf("cannot read the alias file (%s)", err)
		}

		err = yaml.Unmarshal(payload, backend.alias)

		if err != nil {
			backend.alias = map[string]string{}
			return fmt.Errorf("cannot read the manifest content, it is neither a valid JSON nor YAML (%s)", err)
		}
		return nil
	} else {
		return err
	}
}

func (backend *DefaultBackend) extractCmds() {
	for _, repo := range backend.repos {
		cmds := repo.InstalledCommands()
		for _, cmd := range cmds {
			if alias, ok := backend.alias[cmd.FullGroup()]; ok {
				cmd.SetGroupAlias(alias)
			}
			if alias, ok := backend.alias[cmd.FullName()]; ok {
				cmd.SetNameAlias(alias)
			}

			if _, exist := backend.cmdsCache[getCmdSearchKey(cmd)]; exist {
				// conflict
				if cmd.Type() == "group" {
					cmd.SetGroupAlias(cmd.FullGroup())
				} else if cmd.Type() == "executable" {
					if cmd.Group() == "" {
						cmd.SetNameAlias(cmd.FullName())
					} else {
						cmd.SetGroupAlias(cmd.FullGroup())
					}
				}
			}

			key := getCmdSearchKey(cmd)
			backend.cmdsCache[key] = cmd
			switch cmd.Type() {
			case "group":
				backend.groupCmds[key] = cmd
			case "executable":
				backend.executableCmds[key] = cmd
			}
		}
	}
}

func getCmdSearchKey(cmd command.Command) string {
	switch cmd.Type() {
	case "group":
		return fmt.Sprintf("#%s", cmd.GroupOrAlias())
	case "executable":
		return fmt.Sprintf("%s#%s", cmd.GroupOrAlias(), cmd.NameOrAlias())
	case "system":
		return cmd.Name()
	}
	return ""
}

/* Implement the Backend interface */

func (backend *DefaultBackend) FindCommand(group string, name string) (command.Command, error) {
	searchKey := fmt.Sprintf("%s#%s", group, name)
	cmd, ok := backend.cmdsCache[searchKey]
	if !ok {
		return nil, fmt.Errorf("no command with group %s and name %s", group, name)
	}
	return cmd, nil
}

func (backend DefaultBackend) GroupCommands() []command.Command {
	cmds := make([]command.Command, 0)
	for _, v := range backend.groupCmds {
		cmds = append(cmds, v)
	}
	return cmds
}

func (backend DefaultBackend) ExecutableCommands() []command.Command {
	cmds := make([]command.Command, 0)
	for _, v := range backend.executableCmds {
		cmds = append(cmds, v)
	}
	return cmds
}

func (backend *DefaultBackend) RenameCommand(cmd command.Command, new_name string) error {
	if new_name == "" {
		return fmt.Errorf("can't create an empty string alias")
	}

	switch cmd.Type() {
	case "group":
		backend.alias[cmd.FullGroup()] = new_name
	case "executable":
		backend.alias[cmd.FullName()] = new_name
	}

	payload, err := json.Marshal(backend.alias)
	if err != nil {
		return fmt.Errorf("can't encode alias in json: %v", err)
	}

	err = os.WriteFile(filepath.Join(backend.homeDir, "alias.json"), payload, 0755)
	if err != nil {
		return fmt.Errorf("can't write alias filen: %v", err)
	}
	return nil
}

func (backend *DefaultBackend) FindSystemCommand(name string) (command.Command, error) {
	return nil, nil
}

func (backend DefaultBackend) DefaultRepository() repository.PackageRepository {
	return backend.repos[DEFAULT_REPO_INDEX]
}

func (backend DefaultBackend) DropinRepository() repository.PackageRepository {
	return backend.repos[DROPIN_REPO_INDEX]
}

func (backend DefaultBackend) AllRepositories() []repository.PackageRepository {
	return backend.repos
}
