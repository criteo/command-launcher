package repository

import (
	log "github.com/sirupsen/logrus"
)

func CreateLocalRepository(repoDirname string) (PackageRepository, error) {
	repo := newdefaultPackageRepository(repoDirname)
	if err := repo.Load(); err != nil {
		return nil, err
	}

	log.Debug("Local Repository created: ", repo.RepoDir)

	return repo, nil
}

func CreateDropinRepository(repoDirname string) (PackageRepository, error) {
	dropin := newPackageDropin(repoDirname)
	if err := dropin.Load(); err != nil {
		return nil, err
	}

	log.Debug("Dropin Repository created: ", dropin.RepoDir)

	return dropin, nil
}
