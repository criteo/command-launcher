package remote

import (
	"encoding/json"

	"github.com/criteo/command-launcher/internal/command"
)

type PackageInfo struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	Url            string `json:"url"`
	Checksum       string `json:"checksum"`
	StartPartition uint8  `json:"startPartition"`
	EndPartition   uint8  `json:"endPartition"`
}

// Custom unmarshal method to deal with default StartPartition and EndPartition
func (t *PackageInfo) UnmarshalJSON(data []byte) error {
	type packageInfoAlias PackageInfo
	pkgInfo := &packageInfoAlias{
		StartPartition: 0,
		EndPartition:   9,
	}

	err := json.Unmarshal(data, pkgInfo)
	if err != nil {
		return err
	}

	*t = PackageInfo(*pkgInfo)
	return nil
}

// PackagesByVersion type help us to sort the packages by their version
type PackagesByVersion []PackageInfo

func (a PackagesByVersion) Less(i, j int) bool {
	var l defaultVersion
	var r defaultVersion

	err := ParseVersion(a[i].Version, &l)
	if err != nil { // wrong format version is considered smaller
		return true
	}
	err = ParseVersion(a[j].Version, &r)
	if err != nil { // wrong fromat version is considered smaller
		return false
	}

	return Less(l, r)
}
func (a PackagesByVersion) Len() int      { return len(a) }
func (a PackagesByVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// RemoteRepository represents a group of commands packages, and the current version of them
// to be used by cdt
type RemoteRepository interface {
	// Fetch the remote repository metadata
	Fetch() error

	// Get informations of all packages
	All() ([]PackageInfo, error)

	// Get all available package names
	PackageNames() ([]string, error)

	// Get the available versions of a given package
	Versions(packageName string) ([]string, error)

	// Get the latest available version of a given package
	LatestVersion(packageName string) (string, error)

	// Get the latest version of a given package following a filter function
	// this function will pass the package info to the filter from the latest version to the oldest one
	// until the filter returns true
	QueryLatestVersion(packageName string, filter PackageInfoFilterFunc) (string, error)

	// Get the package info of the latest available version
	LatestPackageInfo(packageName string) (*PackageInfo, error)

	// Get the latest package info according to a filter function
	// this function will pass the package to the filter from the latest version to the oldest one
	// until the filter returns true
	QueryLatestPackageInfo(packageName string, filter PackageInfoFilterFunc) (*PackageInfo, error)

	// Download the package of the command with specific version
	Package(packageName string, packageVersion string) (command.Package, error)

	// get the package information of a given version
	PackageInfo(packageName string, packageVersion string) (*PackageInfo, error)

	// Verify package: support two verifications: checksum and signature
	Verify(pkg command.Package, verifyChecksum, verifySignature bool) (bool, error)
}

type PackageInfoFilterFunc func(pkgInfo *PackageInfo) bool
