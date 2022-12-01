package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/criteo/command-launcher/internal/command"
)

type jsonRegistry struct {
	defaultRegistry
	pathname string
}

func newJsonRegistry(path string) (Registry, error) {
	reg := jsonRegistry{
		defaultRegistry: defaultRegistry{
			packages:       make(map[string]command.PackageManifest),
			groupCmds:      make(map[string]command.Command),
			executableCmds: make(map[string]command.Command),
		},
		pathname: path,
	}

	return &reg, nil
}

func (reg *jsonRegistry) Load(repoDir string) error {
	_, err := os.Stat(reg.pathname)
	if !os.IsNotExist(err) {
		payload, err := os.ReadFile(reg.pathname)
		if err != nil {
			return err
		}

		packages := make(map[string]*defaultRegistryEntry, 0)
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

func (reg *jsonRegistry) store() error {
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

func (reg *jsonRegistry) Add(pkg command.PackageManifest, repoDir string) error {
	if err := reg.defaultRegistry.Add(pkg, repoDir); err != nil {
		return err
	}
	return reg.store()
}

func (reg *jsonRegistry) Remove(pkgName string, repoDir string) error {
	if err := reg.defaultRegistry.Remove(pkgName, repoDir); err != nil {
		return err
	}
	return reg.store()
}

func (reg *jsonRegistry) Update(pkg command.PackageManifest, repoDir string) error {
	if err := reg.defaultRegistry.Update(pkg, repoDir); err != nil {
		return err
	}
	return reg.store()
}
