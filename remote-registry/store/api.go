package store

import (
	remote "github.com/criteo/command-launcher/internal/remote"
	model "github.com/criteo/command-launcher/remote-registry/model"
)

type Store interface {
	AllRegistries() ([]model.Registry, error)
	NewRegistry(name string, registryInfo model.RegistryMetadata) error
	UpdateRegistry(name string, registryInfo model.RegistryMetadata) error
	DeleteRegistry(name string) error

	AllPackagesFromRegistry(registryName string) ([]model.Package, error)
	NewPackage(registryName string, packageName string, packageInfo model.PackageMetadata) error
	UpdatePackage(registryName string, packageName string, packageInfo model.PackageMetadata) error
	DeletePackage(registryName string, packageName string) error

	AllPackageVersionsFromRegistry(registryName string, packageName string) ([]remote.PackageInfo, error)
	NewPackageVersion(registryName string, packageName string, version string, packageInfo remote.PackageInfo) error
	DeletePackageVersion(registryName string, packageName string, version string) error
}
