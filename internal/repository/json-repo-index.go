package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/criteo/command-launcher/internal/command"
)

type jsonRepoIndex struct {
	defaultRepoIndex
	pathname string
}

func newJsonRepoIndex(id, path string) (RepoIndex, error) {
	reg := jsonRepoIndex{
		defaultRepoIndex: defaultRepoIndex{
			id:             id,
			packages:       make(map[string]command.PackageManifest),
			packageDirs:    make(map[string]string),
			groupCmds:      make(map[string]command.Command),
			executableCmds: make(map[string]command.Command),
			systemCmds:     make(map[string]command.Command),
		},
		pathname: path,
	}

	return &reg, nil
}

func (reg *jsonRepoIndex) Load(repoDir string) error {
	_, err := os.Stat(reg.pathname)
	if !os.IsNotExist(err) {
		payload, err := os.ReadFile(reg.pathname)
		if err != nil {
			return err
		}

		packages := make(map[string]*defaultRepoIndexEntry, 0)
		if err = json.Unmarshal(payload, &packages); err != nil {
			return err
		}

		for name, pkg := range packages {
			reg.packages[name] = pkg
		}
	}

	reg.extractCmds(repoDir)

	return nil
}

func (reg *jsonRepoIndex) store() error {
	payload, err := json.Marshal(reg.packages)
	if err != nil {
		return fmt.Errorf("cannot encode in json: %v", err)
	}

	err = os.WriteFile(reg.pathname, payload, 0755)
	if err != nil {
		return fmt.Errorf("cannot write registry file: %v", err)
	}

	return nil
}

func (reg *jsonRepoIndex) Add(pkg command.PackageManifest, repoDir string, pkgDirName string) error {
	if err := reg.defaultRepoIndex.Add(pkg, repoDir, pkgDirName); err != nil {
		return err
	}
	return reg.store()
}

func (reg *jsonRepoIndex) Remove(pkgName string, repoDir string) error {
	if err := reg.defaultRepoIndex.Remove(pkgName, repoDir); err != nil {
		return err
	}
	return reg.store()
}

func (reg *jsonRepoIndex) Update(pkg command.PackageManifest, repoDir string, pkgDirName string) error {
	if err := reg.defaultRepoIndex.Update(pkg, repoDir, pkgDirName); err != nil {
		return err
	}
	return reg.store()
}
