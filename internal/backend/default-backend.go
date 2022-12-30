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

// DefaultBackend supports multiple managed repositories and 1 dropin repository
// It contains:
// - 1 dropin local repository - index 0
// - 1 default managed repository - index 1
// - n additional managed repository - index 2 ..
type DefaultBackend struct {
	homeDir string
	sources []*PackageSource
	// repos   []repository.PackageRepository

	cmdsCache      map[string]command.Command
	groupCmds      []command.Command
	executableCmds []command.Command

	userAlias map[string]string
	tmpAlias  map[string]string
}

const DROPIN_REPO_INDEX = 0
const DEFAULT_REPO_INDEX = 1
const DEFAULT_REPO_ID = "default"
const DROPIN_REPO_ID = "dropin"

// Create a new default backend with multiple local repository directories
// When any of these repositories failed to load, an error is returned.
func NewDefaultBackend(homeDir string, dropinSource *PackageSource, defaultSource *PackageSource, additionalSources ...*PackageSource) (Backend, error) {
	backend := &DefaultBackend{
		// input properties
		homeDir: homeDir,
		sources: append([]*PackageSource{dropinSource, defaultSource}, additionalSources...),

		// data need to be reset during reload
		cmdsCache:      map[string]command.Command{},
		groupCmds:      []command.Command{},
		executableCmds: []command.Command{},
		userAlias:      map[string]string{},
		tmpAlias:       map[string]string{},
	}
	err := backend.Reload()
	return backend, err
}

func (backend *DefaultBackend) Reload() error {
	for _, s := range backend.sources {
		s.Repo = nil
	}
	backend.cmdsCache = make(map[string]command.Command)
	backend.groupCmds = []command.Command{}
	backend.executableCmds = []command.Command{}
	backend.userAlias = make(map[string]string)
	backend.tmpAlias = make(map[string]string)

	err := backend.loadRepos()
	backend.loadAlias()
	backend.extractCmds()
	return err
}

func (backend *DefaultBackend) loadRepos() error {
	failures := []string{}
	for i, src := range backend.sources {
		repoDir := src.RepoDir
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
			src.Failure = err
		} else {
			src.Repo = repo
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

		err = yaml.Unmarshal(payload, backend.userAlias)

		if err != nil {
			backend.userAlias = map[string]string{}
			return fmt.Errorf("cannot read the manifest content, it is neither a valid JSON nor YAML (%s)", err)
		}
		return nil
	} else {
		return err
	}
}

func (backend *DefaultBackend) setRuntimeByAlias(cmd command.Command) {
	// first check runtime filter
	if alias, ok := backend.tmpAlias[cmd.FullGroup()]; ok {
		cmd.SetRuntimeGroup(alias)
	}
	if alias, ok := backend.tmpAlias[cmd.FullName()]; ok {
		cmd.SetRuntimeName(alias)
	}
	// override any tmp filer if it defined by user
	if alias, ok := backend.userAlias[cmd.FullGroup()]; ok {
		cmd.SetRuntimeGroup(alias)
	}
	if alias, ok := backend.userAlias[cmd.FullName()]; ok {
		cmd.SetRuntimeName(alias)
	}
}

func (backend *DefaultBackend) extractCmds() {
	for _, src := range backend.sources {
		repo := src.Repo
		if repo == nil {
			continue
		}
		// first extract group commands
		cmds := repo.InstalledGroupCommands()
		for _, cmd := range cmds {
			backend.setRuntimeByAlias(cmd)

			key := getCmdSearchKey(cmd)
			if _, exist := backend.cmdsCache[key]; exist {
				// conflict
				cmd.SetRuntimeName(cmd.FullName())
				backend.tmpAlias[cmd.FullName()] = cmd.FullName()
				key = getCmdSearchKey(cmd)
			}

			backend.cmdsCache[key] = cmd
			backend.groupCmds = append(backend.groupCmds, cmd)
		}

		// now extract executable commands
		cmds = repo.InstalledExecutableCommands()
		for _, cmd := range cmds {
			backend.setRuntimeByAlias(cmd)

			key := getCmdSearchKey(cmd)
			if _, exist := backend.cmdsCache[key]; exist {
				// conflict
				if cmd.Group() == "" {
					cmd.SetRuntimeName(cmd.FullName())
					backend.tmpAlias[cmd.FullName()] = cmd.FullName()
				} else {
					cmd.SetRuntimeGroup(cmd.FullGroup())
					backend.tmpAlias[cmd.FullGroup()] = cmd.FullGroup()
				}
			}

			key = getCmdSearchKey(cmd)
			backend.cmdsCache[key] = cmd
			backend.executableCmds = append(backend.executableCmds, cmd)
		}

		// system commands
		// TODO:
	}
}

func getCmdSearchKey(cmd command.Command) string {
	switch cmd.Type() {
	case "group":
		return fmt.Sprintf("#%s", cmd.RuntimeName())
	case "executable":
		return fmt.Sprintf("%s#%s", cmd.RuntimeGroup(), cmd.RuntimeName())
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
	return backend.groupCmds
}

func (backend DefaultBackend) ExecutableCommands() []command.Command {
	return backend.executableCmds
}

func (backend *DefaultBackend) RenameCommand(cmd command.Command, new_name string) error {
	if new_name == "" {
		return fmt.Errorf("can't create an empty string alias")
	}

	switch cmd.Type() {
	case "group":
		backend.userAlias[cmd.FullGroup()] = new_name
	case "executable":
		backend.userAlias[cmd.FullName()] = new_name
	}

	payload, err := json.Marshal(backend.userAlias)
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
	return backend.sources[DEFAULT_REPO_INDEX].Repo
}

func (backend DefaultBackend) DropinRepository() repository.PackageRepository {
	return backend.sources[DROPIN_REPO_INDEX].Repo
}

func (backend DefaultBackend) AllPackageSources() []*PackageSource {
	return backend.sources
}

func (backend DefaultBackend) AllRepositories() []repository.PackageRepository {
	repos := []repository.PackageRepository{}
	for _, src := range backend.sources {
		repos = append(repos, src.Repo)
	}
	return repos
}

func (backend DefaultBackend) Debug() {
	for _, c := range backend.groupCmds {
		fmt.Printf("%-30s %-30s %s\n", c.RuntimeGroup(), c.RuntimeName(), c.ID())
	}
	for _, c := range backend.executableCmds {
		fmt.Printf("%-30s %-30s %s\n", c.RuntimeGroup(), c.RuntimeName(), c.ID())
	}
	for k, c := range backend.cmdsCache {
		fmt.Printf("%-30s %-30s\n", k, c.ID())
	}
}
