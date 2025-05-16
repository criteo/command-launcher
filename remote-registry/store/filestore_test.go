package store

import (
	"os"
	"path/filepath"
	"testing"

	remote "github.com/criteo/command-launcher/internal/remote"
	model "github.com/criteo/command-launcher/remote-registry/model"
)

func TestFileStore(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filestore-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Initialize empty store", func(t *testing.T) {
		store, err := NewFileStore(tempDir)
		if err != nil {
			t.Fatalf("Failed to create file store: %v", err)
		}

		registries, err := store.AllRegistries()
		if err != nil {
			t.Fatalf("Failed to list registries: %v", err)
		}

		if len(registries) != 0 {
			t.Errorf("Expected empty store, got %d registries", len(registries))
		}
	})

	t.Run("Registry operations", func(t *testing.T) {
		// Create a new store for this test
		storeDir := filepath.Join(tempDir, "registry-ops")
		store, err := NewFileStore(storeDir)
		if err != nil {
			t.Fatalf("Failed to create file store: %v", err)
		}

		// Create registry
		regMeta := model.RegistryMetadata{
			Name:        "test-registry",
			Description: "Test Registry",
			Admin:       []string{"admin"},
			CustomValues: map[string]string{
				"key": "value",
			},
		}

		err = store.NewRegistry("test-registry", regMeta)
		if err != nil {
			t.Fatalf("Failed to create registry: %v", err)
		}

		// Verify registry file exists
		if _, err := os.Stat(filepath.Join(storeDir, "test-registry.json")); os.IsNotExist(err) {
			t.Errorf("Registry file not created")
		}

		// List registries
		registries, err := store.AllRegistries()
		if err != nil {
			t.Fatalf("Failed to list registries: %v", err)
		}

		if len(registries) != 1 {
			t.Errorf("Expected 1 registry, got %d", len(registries))
		}

		if registries[0].Name != "test-registry" {
			t.Errorf("Expected registry name 'test-registry', got '%s'", registries[0].Name)
		}

		// Update registry
		regMeta.Description = "Updated Description"
		err = store.UpdateRegistry("test-registry", regMeta)
		if err != nil {
			t.Fatalf("Failed to update registry: %v", err)
		}

		// Verify update
		registries, _ = store.AllRegistries()
		if registries[0].Description != "Updated Description" {
			t.Errorf("Registry update failed, expected description 'Updated Description', got '%s'",
				registries[0].Description)
		}

		// Delete registry
		err = store.DeleteRegistry("test-registry")
		if err != nil {
			t.Fatalf("Failed to delete registry: %v", err)
		}

		// Verify registry is gone
		registries, _ = store.AllRegistries()
		if len(registries) != 0 {
			t.Errorf("Registry not deleted, still have %d registries", len(registries))
		}

		// Verify file is gone
		if _, err := os.Stat(filepath.Join(storeDir, "test-registry.json")); !os.IsNotExist(err) {
			t.Errorf("Registry file not deleted")
		}
	})

	t.Run("Package operations", func(t *testing.T) {
		// Create a new store for this test
		storeDir := filepath.Join(tempDir, "package-ops")
		store, err := NewFileStore(storeDir)
		if err != nil {
			t.Fatalf("Failed to create file store: %v", err)
		}

		// Create a registry for testing
		regMeta := model.RegistryMetadata{
			Name:        "test-registry",
			Description: "Test Registry",
			Admin:       []string{"admin"},
		}
		store.NewRegistry("test-registry", regMeta)

		// Create package
		pkgMeta := model.PackageMetadata{
			Name:        "test-package",
			Description: "Test Package",
			Admin:       []string{"admin"},
			CustomValues: map[string]string{
				"key": "value",
			},
		}

		err = store.NewPackage("test-registry", "test-package", pkgMeta)
		if err != nil {
			t.Fatalf("Failed to create package: %v", err)
		}

		// List packages
		packages, err := store.AllPackagesFromRegistry("test-registry")
		if err != nil {
			t.Fatalf("Failed to list packages: %v", err)
		}

		if len(packages) != 1 {
			t.Errorf("Expected 1 package, got %d", len(packages))
		}

		if packages[0].Name != "test-package" {
			t.Errorf("Expected package name 'test-package', got '%s'", packages[0].Name)
		}

		// Update package
		pkgMeta.Description = "Updated Package"
		err = store.UpdatePackage("test-registry", "test-package", pkgMeta)
		if err != nil {
			t.Fatalf("Failed to update package: %v", err)
		}

		// Verify update
		packages, _ = store.AllPackagesFromRegistry("test-registry")
		if packages[0].Description != "Updated Package" {
			t.Errorf("Package update failed, expected description 'Updated Package', got '%s'",
				packages[0].Description)
		}

		// Delete package
		err = store.DeletePackage("test-registry", "test-package")
		if err != nil {
			t.Fatalf("Failed to delete package: %v", err)
		}

		// Verify package is gone
		packages, _ = store.AllPackagesFromRegistry("test-registry")
		if len(packages) != 0 {
			t.Errorf("Package not deleted, still have %d packages", len(packages))
		}
	})

	t.Run("Package version operations", func(t *testing.T) {
		// Create a new store for this test
		storeDir := filepath.Join(tempDir, "version-ops")
		store, err := NewFileStore(storeDir)
		if err != nil {
			t.Fatalf("Failed to create file store: %v", err)
		}

		// Create a registry and package for testing
		store.NewRegistry("test-registry", model.RegistryMetadata{Name: "test-registry"})
		store.NewPackage("test-registry", "test-package", model.PackageMetadata{Name: "test-package"})

		// Create package version
		pkgInfo := remote.PackageInfo{
			Name:    "test-package",
			Version: "1.0.0",
			Url:     "http://example.com/package-1.0.0.zip",
		}

		err = store.NewPackageVersion("test-registry", "test-package", "1.0.0", pkgInfo)
		if err != nil {
			t.Fatalf("Failed to create package version: %v", err)
		}

		// List versions
		versions, err := store.AllPackageVersionsFromRegistry("test-registry", "test-package")
		if err != nil {
			t.Fatalf("Failed to list package versions: %v", err)
		}

		if len(versions) != 1 {
			t.Errorf("Expected 1 version, got %d", len(versions))
		}

		if versions[0].Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", versions[0].Version)
		}

		// Add another version
		pkgInfo2 := remote.PackageInfo{
			Name:    "test-package",
			Version: "2.0.0",
			Url:     "http://example.com/package-2.0.0.zip",
		}
		store.NewPackageVersion("test-registry", "test-package", "2.0.0", pkgInfo2)

		// List versions again
		versions, _ = store.AllPackageVersionsFromRegistry("test-registry", "test-package")
		if len(versions) != 2 {
			t.Errorf("Expected 2 versions, got %d", len(versions))
		}

		// Delete version
		err = store.DeletePackageVersion("test-registry", "test-package", "1.0.0")
		if err != nil {
			t.Fatalf("Failed to delete package version: %v", err)
		}

		// Verify version is gone
		versions, _ = store.AllPackageVersionsFromRegistry("test-registry", "test-package")
		if len(versions) != 1 {
			t.Errorf("Version not deleted, still have %d versions", len(versions))
		}

		if versions[0].Version != "2.0.0" {
			t.Errorf("Wrong version deleted, expected '2.0.0', got '%s'", versions[0].Version)
		}
	})

	t.Run("Error conditions", func(t *testing.T) {
		// Create a new store for this test
		storeDir := filepath.Join(tempDir, "error-ops")
		store, err := NewFileStore(storeDir)
		if err != nil {
			t.Fatalf("Failed to create file store: %v", err)
		}

		// Try to update non-existent registry
		err = store.UpdateRegistry("non-existent", model.RegistryMetadata{Name: "non-existent"})
		if err != RegistryDoesNotExistError {
			t.Errorf("Expected RegistryDoesNotExistError, got %v", err)
		}

		// Try to create registry with mismatched name
		err = store.NewRegistry("registry1", model.RegistryMetadata{Name: "registry2"})
		if err != RegistryNameMismatchError {
			t.Errorf("Expected RegistryNameMismatchError, got %v", err)
		}

		// Create a valid registry for further tests
		store.NewRegistry("test-registry", model.RegistryMetadata{Name: "test-registry"})

		// Try to create duplicate registry
		err = store.NewRegistry("test-registry", model.RegistryMetadata{Name: "test-registry"})
		if err != RegistryAlreadyExistsError {
			t.Errorf("Expected RegistryAlreadyExistsError, got %v", err)
		}

		// Try to create package in non-existent registry
		err = store.NewPackage("non-existent", "test-package", model.PackageMetadata{Name: "test-package"})
		if err != RegistryDoesNotExistError {
			t.Errorf("Expected RegistryDoesNotExistError, got %v", err)
		}

		// Create a valid package for further tests
		store.NewPackage("test-registry", "test-package", model.PackageMetadata{Name: "test-package"})

		// Try to create duplicate package
		err = store.NewPackage("test-registry", "test-package", model.PackageMetadata{Name: "test-package"})
		if err != PackageAlreadyExistsError {
			t.Errorf("Expected PackageAlreadyExistsError, got %v", err)
		}

		// Try to create package with mismatched name
		err = store.NewPackage("test-registry", "package1", model.PackageMetadata{Name: "package2"})
		if err != PackageNameMismatchError {
			t.Errorf("Expected PackageNameMismatchError, got %v", err)
		}
	})

	t.Run("Persistence between store instances", func(t *testing.T) {
		// Create a dedicated directory for this test
		storeDir := filepath.Join(tempDir, "persistence")

		// Create first store instance and add data
		store1, err := NewFileStore(storeDir)
		if err != nil {
			t.Fatalf("Failed to create first file store: %v", err)
		}

		// Create registry
		regMeta := model.RegistryMetadata{
			Name:        "persistent-registry",
			Description: "This should persist between store instances",
		}
		store1.NewRegistry("persistent-registry", regMeta)

		// Create package
		pkgMeta := model.PackageMetadata{
			Name:        "persistent-package",
			Description: "This should persist between store instances",
		}
		store1.NewPackage("persistent-registry", "persistent-package", pkgMeta)

		// Create version
		pkgInfo := remote.PackageInfo{
			Name:    "persistent-package",
			Version: "1.0.0",
			Url:     "http://example.com/package.zip",
		}
		store1.NewPackageVersion("persistent-registry", "persistent-package", "1.0.0", pkgInfo)

		// Create a second store instance pointing to same directory
		store2, err := NewFileStore(storeDir)
		if err != nil {
			t.Fatalf("Failed to create second file store: %v", err)
		}

		// Verify registry exists in second instance
		registries, err := store2.AllRegistries()
		if err != nil {
			t.Fatalf("Failed to list registries in second store: %v", err)
		}

		if len(registries) != 1 || registries[0].Name != "persistent-registry" {
			t.Errorf("Registry not persisted between store instances")
		}

		// Verify package exists
		packages, err := store2.AllPackagesFromRegistry("persistent-registry")
		if err != nil {
			t.Fatalf("Failed to list packages in second store: %v", err)
		}

		if len(packages) != 1 || packages[0].Name != "persistent-package" {
			t.Errorf("Package not persisted between store instances")
		}

		// Verify version exists
		versions, err := store2.AllPackageVersionsFromRegistry("persistent-registry", "persistent-package")
		if err != nil {
			t.Fatalf("Failed to list versions in second store: %v", err)
		}

		if len(versions) != 1 || versions[0].Version != "1.0.0" {
			t.Errorf("Package version not persisted between store instances")
		}
	})
}
