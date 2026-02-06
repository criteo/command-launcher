package repository

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/pkg"
)

// workspaceRepoIndex implements RepoIndex for workspace packages.
// Instead of scanning a directory for subdirectories with manifest.mf,
// it loads packages from a pre-supplied list of absolute package directory paths.
type workspaceRepoIndex struct {
	defaultRepoIndex
	packagePaths []string // absolute paths to package folders
}

func NewWorkspaceRepoIndex(id string, packagePaths []string) (RepoIndex, error) {
	base, err := newDefaultRepoIndex(id)
	if err != nil {
		return nil, err
	}

	return &workspaceRepoIndex{
		defaultRepoIndex: *base.(*defaultRepoIndex),
		packagePaths:     packagePaths,
	}, nil
}

// Load ignores the repoDir parameter and loads packages from the pre-supplied paths.
func (idx *workspaceRepoIndex) Load(repoDir string) error {
	for _, pkgPath := range idx.packagePaths {
		manifestPath := filepath.Join(pkgPath, "manifest.mf")
		manifestFile, err := os.Open(manifestPath)
		if err != nil {
			log.Warnf("workspace package: cannot open manifest at %s: %v", manifestPath, err)
			continue
		}

		manifest, err := pkg.ReadManifest(manifestFile)
		manifestFile.Close()
		if err != nil {
			log.Warnf("workspace package: cannot parse manifest at %s: %v", manifestPath, err)
			continue
		}

		idx.packages[manifest.Name()] = manifest
		idx.packageDirs[manifest.Name()] = pkgPath
	}

	idx.extractCmds("")
	return nil
}

// Add returns an error because workspace packages are read-only.
func (idx *workspaceRepoIndex) Add(p command.PackageManifest, repoDir string, pkgDirName string) error {
	return fmt.Errorf("workspace packages are read-only")
}

// Remove returns an error because workspace packages are read-only.
func (idx *workspaceRepoIndex) Remove(pkgName string, repoDir string) error {
	return fmt.Errorf("workspace packages are read-only")
}

// Update returns an error because workspace packages are read-only.
func (idx *workspaceRepoIndex) Update(p command.PackageManifest, repoDir string, pkgDirName string) error {
	return fmt.Errorf("workspace packages are read-only")
}

// IsPackageUpdatePaused always returns false for workspace packages.
func (idx *workspaceRepoIndex) IsPackageUpdatePaused(name string) (bool, error) {
	return false, nil
}

// PausePackageUpdate returns an error because workspace packages are read-only.
func (idx *workspaceRepoIndex) PausePackageUpdate(name string) error {
	return fmt.Errorf("workspace packages are read-only")
}
