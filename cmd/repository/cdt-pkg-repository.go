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

type cdtPackageRepository struct {
	RepoDir  string
	registry *cdtRegistry
}

func newCdtPackageRepository(repoDirname string) *cdtPackageRepository {
	return &cdtPackageRepository{
		RepoDir: repoDirname,
	}
}

func (repo *cdtPackageRepository) Load() error {
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

	log.Debug("Commands loaded: ", repo.registry.AllCommands())
	return nil
}

func (repo *cdtPackageRepository) Install(pkg command.Package) error {
	pkgDir := filepath.Join(repo.RepoDir, pkg.Name())
	err := os.MkdirAll(pkgDir, 0755)
	if err != nil {
		return fmt.Errorf("cannot create the commmand package folder (%v)", err)
	}

	_, err = pkg.InstallTo(pkgDir)
	if err != nil {
		return fmt.Errorf("cannot install the command package %s: %v", pkg.Name(), err)
	}

	err = repo.registry.Add(NewCdtPackage(pkg, pkgDir))
	if err != nil {
		return fmt.Errorf("cannot add the command package %s: %v", pkg.Name(), err)
	}

	err = repo.registry.Store(filepath.Join(repo.RepoDir, FILE_REGISTRY))
	if err != nil {
		return fmt.Errorf("cannot store the new registry %v", err)
	}

	return nil
}

func (repo *cdtPackageRepository) Uninstall(name string) error {
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

func (repo *cdtPackageRepository) Update(pkg command.Package) error {
	err := repo.Uninstall(pkg.Name())
	if err != nil {
		return err
	}

	return repo.Install(pkg)
}

func (repo *cdtPackageRepository) InstalledPackages() []command.PackageManifest {
	return repo.registry.AllPackages()
}

func (repo *cdtPackageRepository) InstalledCommands() []command.Command {
	return repo.registry.AllCommands()
}

func (repo *cdtPackageRepository) InstalledGroupCommands() []command.Command {
	return repo.registry.GroupCommands()
}

func (repo *cdtPackageRepository) InstalledExecutableCommands() []command.Command {
	return repo.registry.ExecutableCommands()
}

func (repo *cdtPackageRepository) Package(name string) (command.PackageManifest, error) {
	return repo.registry.Package(name)
}

func (repo *cdtPackageRepository) Command(group string, name string) (command.Command, error) {
	cmd, err := repo.registry.Command(group, name)
	if err != nil {
		return nil, fmt.Errorf("cannot find the command %s", name)
	}
	return cmd, nil
}
