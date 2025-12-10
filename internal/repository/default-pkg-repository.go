package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/context"
	log "github.com/sirupsen/logrus"
)

const (
	FILE_REGISTRY = "repoIndex.json"
)

type defaultPackageRepository struct {
	ID        string
	RepoDir   string
	repoIndex RepoIndex
}

func newDefaultPackageRepository(id string, repoDirname string, index RepoIndex) *defaultPackageRepository {
	return &defaultPackageRepository{
		ID:        id,
		RepoDir:   repoDirname,
		repoIndex: index,
	}
}

func (repo *defaultPackageRepository) load() error {
	_, err := os.Stat(repo.RepoDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(repo.RepoDir, 0755)
		if err != nil {
			return fmt.Errorf("cannot create the repo folder (%v)", err)
		}
	}

	if err = repo.repoIndex.Load(repo.RepoDir); err != nil {
		return fmt.Errorf("cannot load the packages: %v", err)
	}

	log.Debugf("Commands loaded: %v", func() []string {
		names := []string{}
		for _, cmd := range repo.repoIndex.AllCommands() {
			name := fmt.Sprintf("%s.%s", cmd.Group(), cmd.Name())
			names = append(names, name)
		}
		return names
	}())

	return nil
}

func (repo defaultPackageRepository) Name() string {
	return repo.ID
}

func (repo *defaultPackageRepository) Install(pkg command.Package) error {
	if pkg.Name() == "" {
		return fmt.Errorf("invalid package manifest: empty package name, please make sure manifest.mf contains a 'pkgName'")
	}

	pkgDir := filepath.Join(repo.RepoDir, pkg.Name())

	_, err := pkg.InstallTo(pkgDir)

	// Even if installation fails, we still add the package to the index to pause update
	errAdd := repo.repoIndex.Add(pkg, repo.RepoDir, pkg.Name())
	if errAdd != nil {
		return fmt.Errorf("cannot add the command package %s: %v", pkg.Name(), err)
	}

	// Pause update if installation fails
	if err != nil {
		errPaused := repo.repoIndex.PausePackageUpdate(pkg.Name())
		if errPaused != nil {
			console.Warn("Failed to set update pause for package %s: %v", pkg.Name(), errPaused)
		} else {
			appCtx, _ := context.AppContext()
			console.Reminder(
				"Package %s update has been paused due to installation failure, explicitly run `%s update --package` to retry installation.",
				pkg.Name(),
				appCtx.AppName(),
			)
		}
		errRemove := repo.repoIndex.Remove(pkg.Name(), repo.RepoDir)
		if errRemove != nil {
			console.Warn("Failed to remove package %s: %v", pkg.Name(), errRemove)
		}
		return err
	}

	console.Success("Package %s@%s installed successfully", pkg.Name(), pkg.Version())
	return nil
}

func (repo *defaultPackageRepository) Uninstall(name string) error {
	err := repo.repoIndex.Remove(name, repo.RepoDir)
	if err != nil {
		return fmt.Errorf("cannot remove the command %s: %v", name, err)
	}

	err = os.RemoveAll(filepath.Join(repo.RepoDir, name))
	if err != nil {
		return fmt.Errorf("cannot remove the command folder %v", err)
	}

	return nil
}

func (repo *defaultPackageRepository) Update(pkg command.Package) error {
	err := repo.Uninstall(pkg.Name())
	if err != nil {
		return err
	}

	return repo.Install(pkg)
}

func (repo *defaultPackageRepository) InstalledPackages() []command.PackageManifest {
	return repo.repoIndex.AllPackages()
}

func (repo *defaultPackageRepository) InstalledCommands() []command.Command {
	return repo.repoIndex.AllCommands()
}

func (repo *defaultPackageRepository) InstalledGroupCommands() []command.Command {
	return repo.repoIndex.GroupCommands()
}

func (repo *defaultPackageRepository) InstalledExecutableCommands() []command.Command {
	return repo.repoIndex.ExecutableCommands()
}

func (repo *defaultPackageRepository) InstalledSystemCommands() SystemCommands {
	return SystemCommands{
		Login:   repo.repoIndex.SystemLoginCommand(),
		Metrics: repo.repoIndex.SystemMetricsCommand(),
	}
}

func (repo *defaultPackageRepository) Package(name string) (command.PackageManifest, error) {
	return repo.repoIndex.Package(name)
}

func (repo *defaultPackageRepository) IsPackageUpdatePaused(name string) (bool, error) {
	return repo.repoIndex.IsPackageUpdatePaused(name)
}

func (repo *defaultPackageRepository) PausePackageUpdate(name string) error {
	return repo.repoIndex.PausePackageUpdate(name)
}

func (repo *defaultPackageRepository) Command(pkg string, group string, name string) (command.Command, error) {
	cmd, err := repo.repoIndex.Command(pkg, group, name)
	if err != nil {
		return nil, fmt.Errorf("cannot find the command %s", name)
	}
	return cmd, nil
}

func (repo *defaultPackageRepository) RepositoryFolder() (string, error) {
	return repo.RepoDir, nil
}
