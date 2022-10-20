package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/criteo/command-launcher/internal/command"
	log "github.com/sirupsen/logrus"
)

const (
	FILE_REGISTRY = "registry.json"
)

type defaultPackageRepository struct {
	RepoDir  string
	registry *defaultRegistry
}

func newdefaultPackageRepository(repoDirname string) *defaultPackageRepository {
	return &defaultPackageRepository{
		RepoDir: repoDirname,
	}
}

func (repo *defaultPackageRepository) Load() error {
	_, err := os.Stat(repo.RepoDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(repo.RepoDir, 0755)
		if err != nil {
			return fmt.Errorf("cannot create the repo folder (%v)", err)
		}
	}

	pathname := filepath.Join(repo.RepoDir, FILE_REGISTRY)
	reg, err := LoadRegistry(pathname)
	if err != nil {
		return err
	}

	repo.registry = reg

	log.Debugf("Commands loaded: %v", func() []string {
		names := []string{}
		for _, cmd := range repo.registry.AllCommands() {
			name := fmt.Sprintf("%s.%s", cmd.Group(), cmd.Name())
			names = append(names, name)
		}
		return names
	}())
	return nil
}

func (repo *defaultPackageRepository) Install(pkg command.Package) error {
	if pkg.Name() == "" {
		return fmt.Errorf("invalid package manifest: empty package name, please make sure manifest.mf contains a 'pkgName'")
	}

	pkgDir := filepath.Join(repo.RepoDir, pkg.Name())
	err := os.MkdirAll(pkgDir, 0755)
	if err != nil {
		return fmt.Errorf("cannot create the commmand package folder (%v)", err)
	}

	_, err = pkg.InstallTo(pkgDir)
	if err != nil {
		return fmt.Errorf("cannot install the command package %s: %v", pkg.Name(), err)
	}

	err = repo.registry.Add(newRegistryEntry(pkg, pkgDir))
	if err != nil {
		return fmt.Errorf("cannot add the command package %s: %v", pkg.Name(), err)
	}

	err = repo.registry.Store(filepath.Join(repo.RepoDir, FILE_REGISTRY))
	if err != nil {
		return fmt.Errorf("cannot store the new registry %v", err)
	}

	return nil
}

func (repo *defaultPackageRepository) Uninstall(name string) error {
	err := repo.registry.Remove(name)
	if err != nil {
		return fmt.Errorf("cannot remove the command %s: %v", name, err)
	}

	err = repo.registry.Store(filepath.Join(repo.RepoDir, FILE_REGISTRY))
	if err != nil {
		return fmt.Errorf("cannot store the new registry %v", err)
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
	return repo.registry.AllPackages()
}

func (repo *defaultPackageRepository) InstalledCommands() []command.Command {
	return repo.registry.AllCommands()
}

func (repo *defaultPackageRepository) InstalledGroupCommands() []command.Command {
	return repo.registry.GroupCommands()
}

func (repo *defaultPackageRepository) InstalledExecutableCommands() []command.Command {
	return repo.registry.ExecutableCommands()
}

func (repo *defaultPackageRepository) Package(name string) (command.PackageManifest, error) {
	return repo.registry.Package(name)
}

func (repo *defaultPackageRepository) Command(group string, name string) (command.Command, error) {
	cmd, err := repo.registry.Command(group, name)
	if err != nil {
		return nil, fmt.Errorf("cannot find the command %s", name)
	}
	return cmd, nil
}
