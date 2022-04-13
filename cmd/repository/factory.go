package repository

import (
	log "github.com/sirupsen/logrus"
)

func CreateLocalRepository(repoDirname string) (PackageRepository, error) {
	repo := newdefaultPackageRepository(repoDirname)
	if err := repo.Load(); err != nil {
		return nil, err
	}

	log.Debug("Repository created: ", repo)

	return repo, nil
}
