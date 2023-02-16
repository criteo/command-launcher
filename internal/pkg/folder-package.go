package pkg

import (
	"os"
	"path/filepath"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/spf13/viper"
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

	if viper.GetBool(config.ENABLE_PACKAGE_SETUP_HOOK_KEY) {
		// for now ignore the setup error
		pkg.RunSetup(dstDir)
	}

	return pkg.Manifest, nil
}

func (pkg *folderPackage) VerifyChecksum(checksum string) (bool, error) {
	// TODO: what is the checksum for a folder package?
	return true, nil
}

func (pkg *folderPackage) VerifySignature(signature string) (bool, error) {
	// TODO: what is the signature for a folder package?
	return true, nil
}
