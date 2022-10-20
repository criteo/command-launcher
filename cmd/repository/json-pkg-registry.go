package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/criteo/command-launcher/internal/command"
)

type jsonRegistry struct {
	defaultRegistry
	pathname string
}

func newJsonRegistry(path string) (*Registry, error) {
	reg := jsonRegistry{
		defaultRegistry{
			packages:       make(map[string]defaultRegistryEntry),
			groupCmds:      make(map[string]*command.DefaultCommand),
			executableCmds: make(map[string]*command.DefaultCommand),
		},
		pathname: path,
	}

	err := reg.load()

	return &reg, err
}

func (reg *jsonRegistry) load() error {
	_, err := os.Stat(reg.pathname)
	if !os.IsNotExist(err) {
		payload, err := ioutil.ReadFile(reg.pathname)
		if err != nil {
			return err
		}

		err = json.Unmarshal(payload, &reg.defaultRegistry.packages)
		if err != nil {
			return err
		}
	}

	reg.defaultRegistry.extractCmds()

	return nil
}

func (reg *jsonRegistry) store() error {
	payload, err := json.Marshal(reg.defaultRegistry.packages)
	if err != nil {
		return fmt.Errorf("cannot encode in json: %v", err)
	}

	err = ioutil.WriteFile(reg.pathname, payload, 0755)
	if err != nil {
		return fmt.Errorf("cannot write registry file: %v", err)
	}

	return nil
}

func (reg *jsonRegistry) Add(pkg command.PackageManifest) error {
	if err := reg.defaultRegistry.Add(pkg); err != nil {
		return err
	}

	return reg.store()
}

func (reg *jsonRegistry) Remove(name string) (command.PackageManifest, error) {
	pkg, err := reg.defaultRegistry.Remove(name)
	if err != nil {
		return nil, err
	}

	err = reg.store()
	return pkg, err
}

func (reg *jsonRegistry) Update(pkg command.PackageManifest) (command.PackageManifest, error) {
	pkg, err := reg.defaultRegistry.Update(pkg)
	if err != nil {
		return nil, err
	}

	err = reg.store()
	return pkg, err
}
