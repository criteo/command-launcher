package repository

import (
	log "github.com/sirupsen/logrus"
)

func CreateLocalRepository(repoDirname string, registry Registry) (PackageRepository, error) {
	var reg Registry
	if registry == nil {
		var err error
		reg, err = newDefaultRegistry()
		if err != nil {
			return nil, err
		}
	} else {
		reg = registry
	}

	repo := newdefaultPackageRepository(repoDirname, reg)
	if err := repo.load(); err != nil {
		return nil, err
	}

	log.Debugf("Repository created, Folder %s", repo.RepoDir)

	return repo, nil
}
