package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	remote "github.com/criteo/command-launcher/internal/remote"
	model "github.com/criteo/command-launcher/remote-registry/model"
)

// FileStore implements the storage interface with file system persistence
type FileStore struct {
	basePath string
	mutex    sync.RWMutex
	// Cache to avoid reading from disk on every operation
	registriesCache map[string]*model.Registry
}

// NewFileStore creates a new file-based store
func NewFileStore(storagePath string) (*FileStore, error) {
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	store := &FileStore{
		basePath:        storagePath,
		registriesCache: make(map[string]*model.Registry),
	}

	// Initialize cache by loading existing registries
	if err := store.loadRegistriesFromDisk(); err != nil {
		return nil, err
	}

	return store, nil
}

// loadRegistriesFromDisk reads all registry files into cache
func (s *FileStore) loadRegistriesFromDisk() error {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return fmt.Errorf("failed to read storage directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		registryName := entry.Name()[:len(entry.Name())-5] // Remove .json
		registry, err := s.loadRegistryFromDisk(registryName)
		if err != nil {
			return err
		}

		s.registriesCache[registryName] = registry
	}

	return nil
}

// loadRegistryFromDisk loads a single registry from disk
func (s *FileStore) loadRegistryFromDisk(name string) (*model.Registry, error) {
	path := filepath.Join(s.basePath, fmt.Sprintf("%s.json", name))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, RegistryDoesNotExistError
		}
		return nil, fmt.Errorf("failed to read registry file: %w", err)
	}

	var registry model.Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry file: %w", err)
	}

	return &registry, nil
}

// saveRegistryToDisk writes a registry to disk
func (s *FileStore) saveRegistryToDisk(registry *model.Registry) error {
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	path := filepath.Join(s.basePath, fmt.Sprintf("%s.json", registry.Name))
	return os.WriteFile(path, data, 0644)
}

// NewRegistry creates a new registry
func (s *FileStore) NewRegistry(name string, registryInfo model.RegistryMetadata) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.registriesCache[name]; exists {
		return RegistryAlreadyExistsError
	}
	if name != registryInfo.Name {
		return RegistryNameMismatchError
	}

	registry := &model.Registry{
		RegistryMetadata: registryInfo,
		Packages:         make(map[string]model.Package),
	}

	// Save to disk
	if err := s.saveRegistryToDisk(registry); err != nil {
		return err
	}

	// Update cache
	s.registriesCache[name] = registry
	return nil
}

// UpdateRegistry updates an existing registry
func (s *FileStore) UpdateRegistry(name string, registryInfo model.RegistryMetadata) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[name]
	if !exists {
		return RegistryDoesNotExistError
	}
	if name != registryInfo.Name {
		return RegistryNameMismatchError
	}

	// Update registry metadata
	registry.Description = registryInfo.Description
	registry.Admin = registryInfo.Admin
	registry.CustomValues = registryInfo.CustomValues

	// Save to disk
	if err := s.saveRegistryToDisk(registry); err != nil {
		return err
	}

	return nil
}

// DeleteRegistry removes a registry
func (s *FileStore) DeleteRegistry(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.registriesCache[name]; !exists {
		return RegistryDoesNotExistError
	}

	// Remove from disk
	path := filepath.Join(s.basePath, fmt.Sprintf("%s.json", name))
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete registry file: %w", err)
	}

	// Remove from cache
	delete(s.registriesCache, name)
	return nil
}

// AllRegistries returns all registries
func (s *FileStore) AllRegistries() ([]model.Registry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	registries := []model.Registry{}
	for _, registry := range s.registriesCache {
		registries = append(registries, *registry)
	}

	// Sort registries by name
	sort.Slice(registries, func(i, j int) bool {
		return registries[i].Name < registries[j].Name
	})

	return registries, nil
}

// NewPackage creates a new package in a registry
func (s *FileStore) NewPackage(registryName string, packageName string, packageInfo model.PackageMetadata) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return RegistryDoesNotExistError
	}
	if _, exists := registry.Packages[packageName]; exists {
		return PackageAlreadyExistsError
	}
	if packageName != packageInfo.Name {
		return PackageNameMismatchError
	}

	// Add package
	registry.Packages[packageName] = model.Package{
		PackageMetadata: packageInfo,
		Versions:        []remote.PackageInfo{},
	}

	// Save to disk
	if err := s.saveRegistryToDisk(registry); err != nil {
		return err
	}

	return nil
}

// UpdatePackage updates an existing package
func (s *FileStore) UpdatePackage(registryName string, packageName string, packageInfo model.PackageMetadata) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return RegistryDoesNotExistError
	}
	pkg, exists := registry.Packages[packageName]
	if !exists {
		return PackageDoesNotExistError
	}
	if packageName != packageInfo.Name {
		return PackageNameMismatchError
	}

	// Update package metadata
	pkg.Description = packageInfo.Description
	pkg.Admin = packageInfo.Admin
	pkg.CustomValues = packageInfo.CustomValues
	registry.Packages[packageName] = pkg

	// Save to disk
	if err := s.saveRegistryToDisk(registry); err != nil {
		return err
	}

	return nil
}

// DeletePackage removes a package from a registry
func (s *FileStore) DeletePackage(registryName string, packageName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return RegistryDoesNotExistError
	}
	if _, exists := registry.Packages[packageName]; !exists {
		return PackageDoesNotExistError
	}

	// Delete package
	delete(registry.Packages, packageName)

	// Save to disk
	if err := s.saveRegistryToDisk(registry); err != nil {
		return err
	}

	return nil
}

// AllPackagesFromRegistry returns all packages in a registry
func (s *FileStore) AllPackagesFromRegistry(registryName string) ([]model.Package, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return nil, RegistryDoesNotExistError
	}

	packages := []model.Package{}
	for _, pkg := range registry.Packages {
		packages = append(packages, pkg)
	}

	// Sort packages by name
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	return packages, nil
}

// NewPackageVersion adds a version to a package
func (s *FileStore) NewPackageVersion(registryName string, packageName string, version string, packageInfo remote.PackageInfo) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return RegistryDoesNotExistError
	}
	pkg, exists := registry.Packages[packageName]
	if !exists {
		return PackageDoesNotExistError
	}
	if packageName != packageInfo.Name {
		return PackageNameMismatchError
	}

	// Check if version already exists
	for _, v := range pkg.Versions {
		if v.Version == version {
			return PackageVersionAlreadyExistsError
		}
	}

	// Add version
	pkg.Versions = append(pkg.Versions, packageInfo)
	registry.Packages[packageName] = pkg

	// Save to disk
	if err := s.saveRegistryToDisk(registry); err != nil {
		return err
	}

	return nil
}

// DeletePackageVersion removes a version from a package
func (s *FileStore) DeletePackageVersion(registryName string, packageName string, version string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return RegistryDoesNotExistError
	}
	pkg, exists := registry.Packages[packageName]
	if !exists {
		return PackageDoesNotExistError
	}

	// Find and remove version
	newVersions := []remote.PackageInfo{}
	exists = false
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
	registry.Packages[packageName] = pkg

	// Save to disk
	if err := s.saveRegistryToDisk(registry); err != nil {
		return err
	}

	return nil
}

// AllPackageVersionsFromRegistry returns all versions of a package
func (s *FileStore) AllPackageVersionsFromRegistry(registryName string, packageName string) ([]remote.PackageInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return nil, RegistryDoesNotExistError
	}
	pkg, exists := registry.Packages[packageName]
	if !exists {
		return nil, PackageDoesNotExistError
	}

	versions := make([]remote.PackageInfo, len(pkg.Versions))
	copy(versions, pkg.Versions)

	return versions, nil
}
