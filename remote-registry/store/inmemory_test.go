package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	remote "github.com/criteo/command-launcher/internal/remote"
	model "github.com/criteo/command-launcher/remote-registry/model"
)

func TestInMemoryStore_Registry_CRUD(t *testing.T) {
	store := NewInMemoryStore()

	// Test NewRegistry
	registryInfo := model.RegistryMetadata{
		Name:         "test-registry",
		Description:  "Test Registry",
		Admin:        []string{"a", "b"},
		CustomValues: map[string]string{"key": "value"},
	}

	store.NewRegistry("test-registry", registryInfo)

	allRegistries, err := store.AllRegistries()

	assert.NoError(t, err)
	assert.Len(t, allRegistries, 1)
	assert.Equal(t, "test-registry", allRegistries[0].Name)
	assert.Equal(t, "Test Registry", allRegistries[0].Description)
	assert.Equal(t, []string{"a", "b"}, allRegistries[0].Admin)
	assert.Equal(t, map[string]string{"key": "value"}, allRegistries[0].CustomValues)

	// Add another registry
	registryInfo2 := model.RegistryMetadata{
		Name:         "test-registry-2",
		Description:  "Test Registry 2",
		Admin:        []string{"c", "d"},
		CustomValues: map[string]string{"key2": "value2"},
	}
	store.NewRegistry("test-registry-2", registryInfo2)
	allRegistries, err = store.AllRegistries()
	assert.NoError(t, err)
	assert.Len(t, allRegistries, 2)
	assert.Equal(t, "test-registry-2", allRegistries[1].Name)
	assert.Equal(t, "Test Registry 2", allRegistries[1].Description)
	assert.Equal(t, []string{"c", "d"}, allRegistries[1].Admin)
	assert.Equal(t, map[string]string{"key2": "value2"}, allRegistries[1].CustomValues)

	// add a registry with same name
	err = store.NewRegistry("test-registry", registryInfo)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, RegistryAlreadyExistsError))

	// Update Registry
	registryInfo.Description = "Updated Test Registry"
	err = store.UpdateRegistry("test-registry", registryInfo)
	assert.NoError(t, err)
	allRegistries, err = store.AllRegistries()
	assert.NoError(t, err)
	assert.Len(t, allRegistries, 2)
	assert.Equal(t, "Updated Test Registry", allRegistries[0].Description)
	assert.Equal(t, "Test Registry 2", allRegistries[1].Description)

	// Delete Registry
	err = store.DeleteRegistry("test-registry")
	assert.NoError(t, err)
	allRegistries, err = store.AllRegistries()
	assert.NoError(t, err)
	assert.Len(t, allRegistries, 1)
	assert.Equal(t, "test-registry-2", allRegistries[0].Name)
	assert.Equal(t, "Test Registry 2", allRegistries[0].Description)
}

func TestInMemoryStore_Package_CRUD_Errors(t *testing.T) {
	store := NewInMemoryStore()
	registryInfo := model.RegistryMetadata{
		Name:         "test-registry",
		Description:  "Test Registry",
		Admin:        []string{"a", "b"},
		CustomValues: map[string]string{"key": "value"},
	}

	store.NewRegistry("test-registry", registryInfo)
	packageInfo := model.PackageMetadata{
		Name:         "test-package",
		Description:  "Test Package",
		Admin:        []string{"a", "b"},
		CustomValues: map[string]string{"key": "value"},
	}

	error := store.NewPackage("test-registry", "test-package", packageInfo)
	assert.NoError(t, error)

	allPackages, error := store.AllPackagesFromRegistry("test-registry")
	assert.NoError(t, error)
	assert.Len(t, allPackages, 1)
	assert.Equal(t, "test-package", allPackages[0].Name)

	// Add another package
	packageInfo2 := model.PackageMetadata{
		Name:         "test-package-2",
		Description:  "Test Package 2",
		Admin:        []string{"c", "d"},
		CustomValues: map[string]string{"key2": "value2"},
	}
	error = store.NewPackage("test-registry", "test-package-2", packageInfo2)
	assert.NoError(t, error)
	allPackages, error = store.AllPackagesFromRegistry("test-registry")
	assert.NoError(t, error)
	assert.Len(t, allPackages, 2)
	assert.Equal(t, "test-package-2", allPackages[1].Name)
	assert.Equal(t, "Test Package 2", allPackages[1].Description)

	// Update Package
	packageInfo.Description = "Updated Test Package"
	error = store.UpdatePackage("test-registry", "test-package", packageInfo)
	assert.NoError(t, error)
	allPackages, error = store.AllPackagesFromRegistry("test-registry")
	assert.NoError(t, error)
	assert.Len(t, allPackages, 2)
	assert.Equal(t, "Updated Test Package", allPackages[0].Description)
	assert.Equal(t, "Test Package 2", allPackages[1].Description)

	// Delete Package
	error = store.DeletePackage("test-registry", "test-package")
	assert.NoError(t, error)
	allPackages, error = store.AllPackagesFromRegistry("test-registry")
	assert.NoError(t, error)
	assert.Len(t, allPackages, 1)
	assert.Equal(t, "test-package-2", allPackages[0].Name)
	assert.Equal(t, "Test Package 2", allPackages[0].Description)
}

func TestInMemoryStore_PackageVersion_CRUD_Errors(t *testing.T) {
	store := NewInMemoryStore()
	registryInfo := model.RegistryMetadata{
		Name:         "test-registry",
		Description:  "Test Registry",
		Admin:        []string{"a", "b"},
		CustomValues: map[string]string{"key": "value"},
	}
	store.NewRegistry("test-registry", registryInfo)
	packageInfo := model.PackageMetadata{
		Name:         "test-package",
		Description:  "Test Package",
		Admin:        []string{"a", "b"},
		CustomValues: map[string]string{"key": "value"},
	}
	store.NewPackage("test-registry", "test-package", packageInfo)

	packageVersionInfo := remote.PackageInfo{
		Name:           "test-package",
		Version:        "1.0.0",
		Url:            "http://example.com/test-package-1.0.0.tar.gz",
		Checksum:       "abc123",
		StartPartition: 0,
		EndPartition:   9,
	}

	error := store.NewPackageVersion("test-registry", "test-package", "1.0.0", packageVersionInfo)
	assert.NoError(t, error)
	allPackageVersions, error := store.AllPackageVersionsFromRegistry("test-registry", "test-package")
	assert.NoError(t, error)
	assert.Len(t, allPackageVersions, 1)
	assert.Equal(t, "test-package", allPackageVersions[0].Name)
	assert.Equal(t, "1.0.0", allPackageVersions[0].Version)
	assert.Equal(t, "http://example.com/test-package-1.0.0.tar.gz", allPackageVersions[0].Url)

	// Add another package version
	packageVersionInfo2 := remote.PackageInfo{
		Name:           "test-package",
		Version:        "2.0.0",
		Url:            "http://example.com/test-package-2.0.0.tar.gz",
		Checksum:       "def456",
		StartPartition: 0,
		EndPartition:   9,
	}
	error = store.NewPackageVersion("test-registry", "test-package", "2.0.0", packageVersionInfo2)
	assert.NoError(t, error)
	allPackageVersions, error = store.AllPackageVersionsFromRegistry("test-registry", "test-package")
	assert.NoError(t, error)
	assert.Len(t, allPackageVersions, 2)
	assert.Equal(t, "test-package", allPackageVersions[1].Name)
	assert.Equal(t, "2.0.0", allPackageVersions[1].Version)
	assert.Equal(t, "http://example.com/test-package-2.0.0.tar.gz", allPackageVersions[1].Url)

	// Add a package version with different package name
	packageVersionInfo3 := remote.PackageInfo{
		Name:           "test-package-2",
		Version:        "3.0.0",
		Url:            "http://example.com/test-package-3.0.0.tar.gz",
		Checksum:       "ghi789",
		StartPartition: 0,
		EndPartition:   9,
	}
	error = store.NewPackageVersion("test-registry", "test-package", "3.0.0", packageVersionInfo3)
	assert.Error(t, error)
	assert.True(t, errors.Is(error, PackageNameMismatchError))

	// Delete Package Version
	error = store.DeletePackageVersion("test-registry", "test-package", "1.0.0")
	assert.NoError(t, error)
	allPackageVersions, error = store.AllPackageVersionsFromRegistry("test-registry", "test-package")
	assert.NoError(t, error)
	assert.Len(t, allPackageVersions, 1)
	assert.Equal(t, "test-package", allPackageVersions[0].Name)
	assert.Equal(t, "2.0.0", allPackageVersions[0].Version)

	// Delete Package Version that does not exist
	error = store.DeletePackageVersion("test-registry", "test-package", "1.0.0")
	assert.Error(t, error)
	assert.True(t, errors.Is(error, PackageVersionDoesNotExistError))
}
