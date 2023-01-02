package repository

import (
	log "github.com/sirupsen/logrus"
)

func CreateLocalRepository(id string, repoDirname string, idx RepoIndex) (PackageRepository, error) {
	var repoIndex RepoIndex
	if idx == nil {
		var err error
		repoIndex, err = newDefaultRepoIndex(id)
		if err != nil {
			return nil, err
		}
	} else {
		repoIndex = idx
	}

	repo := newDefaultPackageRepository(id, repoDirname, repoIndex)
	if err := repo.load(); err != nil {
		return nil, err
	}

	log.Debugf("Repository created, Folder %s", repo.RepoDir)

	return repo, nil
}
