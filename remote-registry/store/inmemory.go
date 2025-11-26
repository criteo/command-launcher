package store

import (
	"sort"

	remote "github.com/criteo/command-launcher/internal/remote"
	model "github.com/criteo/command-launcher/remote-registry/model"
)

type InMemoryStore struct {
	registries map[string]model.Registry
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		registries: make(map[string]model.Registry),
	}
}

func (s *InMemoryStore) NewRegistry(name string, registryInfo model.RegistryMetadata) error {
	if _, exists := s.registries[name]; exists {
		return RegistryAlreadyExistsError
	}
	if name != registryInfo.Name {
		return RegistryNameMismatchError
	}
	s.registries[name] = model.Registry{
		RegistryMetadata: registryInfo,
		Packages:         make(map[string]model.Package),
	}
	return nil
}

func (s *InMemoryStore) UpdateRegistry(name string, registryInfo model.RegistryMetadata) error {
	if _, exists := s.registries[name]; !exists {
		return RegistryDoesNotExistError
	}
	if name != registryInfo.Name {
		return RegistryNameMismatchError
	}

	registry := s.registries[name]
	registry.Description = registryInfo.Description
	registry.Admin = registryInfo.Admin
	registry.CustomValues = registryInfo.CustomValues

	s.registries[name] = registry

	return nil
}

func (s *InMemoryStore) DeleteRegistry(name string) error {
	if _, exists := s.registries[name]; !exists {
		return RegistryDoesNotExistError
	}
	delete(s.registries, name)
	return nil
}

func (s InMemoryStore) AllRegistries() ([]model.Registry, error) {
	registries := []model.Registry{}
	for _, registry := range s.registries {
		registries = append(registries, registry)
	}

	// sort registries by name
	sort.Slice(registries, func(i, j int) bool {
		return registries[i].Name < registries[j].Name
	})

	return registries, nil
}

func (s *InMemoryStore) NewPackage(registryName string, packageName string, packageInfo model.PackageMetadata) error {
	if _, exists := s.registries[registryName]; !exists {
		return RegistryDoesNotExistError
	}
	if _, exists := s.registries[registryName].Packages[packageName]; exists {
		return PackageAlreadyExistsError
	}
	if packageName != packageInfo.Name {
		return PackageNameMismatchError
	}
	s.registries[registryName].Packages[packageName] = model.Package{
		PackageMetadata: packageInfo,
		Versions:        []remote.PackageInfo{},
	}
	return nil
}

func (s *InMemoryStore) UpdatePackage(registryName string, packageName string, packageInfo model.PackageMetadata) error {
	if _, exists := s.registries[registryName]; !exists {
		return RegistryDoesNotExistError
	}
	if _, exists := s.registries[registryName].Packages[packageName]; !exists {
		return PackageDoesNotExistError
	}
	if packageName != packageInfo.Name {
		return PackageNameMismatchError
	}

	pkg := s.registries[registryName].Packages[packageName]
	pkg.Description = packageInfo.Description
	pkg.Admin = packageInfo.Admin
	pkg.CustomValues = packageInfo.CustomValues

	s.registries[registryName].Packages[packageName] = pkg

	return nil
}

func (s *InMemoryStore) DeletePackage(registryName string, packageName string) error {
	if _, exists := s.registries[registryName]; !exists {
		return RegistryDoesNotExistError
	}
	if _, exists := s.registries[registryName].Packages[packageName]; !exists {
		return PackageDoesNotExistError
	}
	delete(s.registries[registryName].Packages, packageName)
	return nil
}

func (s InMemoryStore) AllPackagesFromRegistry(registryName string) ([]model.Package, error) {
	if _, exists := s.registries[registryName]; !exists {
		return nil, RegistryDoesNotExistError
	}
	packages := []model.Package{}
	for _, pkg := range s.registries[registryName].Packages {
		packages = append(packages, pkg)
	}

	// sort packages by name
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	return packages, nil
}

func (s *InMemoryStore) NewPackageVersion(registryName string, packageName string, version string, packageInfo remote.PackageInfo) error {
	if _, exists := s.registries[registryName]; !exists {
		return RegistryDoesNotExistError
	}
	if _, exists := s.registries[registryName].Packages[packageName]; !exists {
		return PackageDoesNotExistError
	}

	if packageName != packageInfo.Name {
		return PackageNameMismatchError
	}

	pkg := s.registries[registryName].Packages[packageName]
	// Check if the version already exists
	for _, v := range pkg.Versions {
		if v.Version == version {
			return PackageVersionAlreadyExistsError
		}
	}
	pkg.Versions = append(pkg.Versions, packageInfo)
	s.registries[registryName].Packages[packageName] = pkg

	return nil
}

func (s *InMemoryStore) DeletePackageVersion(registryName string, packageName string, version string) error {
	if _, exists := s.registries[registryName]; !exists {
		return RegistryDoesNotExistError
	}
	if _, exists := s.registries[registryName].Packages[packageName]; !exists {
		return PackageDoesNotExistError
	}

	pkg := s.registries[registryName].Packages[packageName]
	newVersions := []remote.PackageInfo{}
	exists := false
	for _, v := range pkg.Versions {
		if v.Version == version {
			exists = true
		}
		if v.Version != version {
			newVersions = append(newVersions, v)
		}
	}
	if !exists {
		return PackageVersionDoesNotExistError
	}
	pkg.Versions = newVersions

	s.registries[registryName].Packages[packageName] = pkg

	return nil
}

func (s InMemoryStore) AllPackageVersionsFromRegistry(registryName string, packageName string) ([]remote.PackageInfo, error) {
	if _, exists := s.registries[registryName]; !exists {
		return nil, RegistryDoesNotExistError
	}
	if _, exists := s.registries[registryName].Packages[packageName]; !exists {
		return nil, PackageDoesNotExistError
	}
	versions := []remote.PackageInfo{}
	for _, v := range s.registries[registryName].Packages[packageName].Versions {
		versions = append(versions, v)
	}

	return versions, nil
}
