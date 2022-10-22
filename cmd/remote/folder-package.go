package remote

import (
	"os"
	"path/filepath"

	"github.com/criteo/command-launcher/internal/command"
)

type folderPackage struct {
	defaultPackage
	sourceDir string
}

func CreateFolderPackage(folder string) (command.Package, error) {
	manifestFile, err := os.Open(filepath.Join(folder, "manifest.mf"))
	if err != nil {
		return nil, err
	}

	manifest, err := ReadManifest(manifestFile)
	if err != nil {
		return nil, err
	}

	pkg := folderPackage{
		defaultPackage: defaultPackage{
			Manifest: manifest,
		},
		sourceDir: folder,
	}

	return &pkg, nil
}

func (pkg *folderPackage) InstallTo(targetDir string) (command.PackageManifest, error) {
	dstDir := filepath.Join(targetDir, pkg.Manifest.Name())
	if err := copyFolder(pkg.sourceDir, dstDir); err != nil {
		return nil, err
	}

	return pkg.Manifest, nil
}
