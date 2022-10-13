package repository

import (
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/criteo/command-launcher/cmd/remote"
)

type defaultDropinRepository struct {
	defaultPackageRepository
}

func newPackageDropin(repoDirname string) *defaultDropinRepository {
	return &defaultDropinRepository{
		defaultPackageRepository{
			RepoDir:  repoDirname,
			registry: newRegistry(),
		},
	}
}

func (dropin *defaultDropinRepository) Load() error {
	_, err := os.Stat(dropin.RepoDir)
	if !os.IsNotExist(err) {
		files, err := ioutil.ReadDir(dropin.RepoDir)
		if err != nil {
			log.Errorf("cannot read dropin folder: %v", err)
			return err
		}

		for _, f := range files {
			if !f.IsDir() && f.Mode()&os.ModeSymlink != os.ModeSymlink {
				continue
			}

			pkgFolder := filepath.Join(dropin.RepoDir, f.Name())
			pkg, err := remote.CreateFolderPackage(pkgFolder)
			if err == nil {
				dropin.registry.Add(NewRegistryEntry(pkg, pkgFolder))
			}
		}
	}

	return nil
}
