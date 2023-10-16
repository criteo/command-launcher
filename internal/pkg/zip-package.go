package pkg

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type zipPackage struct {
	defaultPackage
	ZipFile string
}

func CreateZipPackage(zipFilename string) (command.Package, error) {
	reader, err := zip.OpenReader(zipFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open: %s", err)
	}
	defer reader.Close()
	manifestFile, err := reader.Open("manifest.mf")
	if err != nil {
		return nil, fmt.Errorf("failed to open the manifest: %s", err)
	}
	defer manifestFile.Close()

	mf, err := ReadManifest(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read the manifest: %s", err)
	}

	var pkg = zipPackage{
		defaultPackage: defaultPackage{
			Manifest: mf,
		},
		ZipFile: zipFilename,
	}

	return &pkg, nil
}

func (pkg *zipPackage) InstallTo(targetDir string) (command.PackageManifest, error) {
	zipReader, _ := zip.OpenReader(pkg.ZipFile)
	defer zipReader.Close()
	for _, file := range zipReader.Reader.File {
		zippedFile, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("installation failed: %s", err)
		}
		defer zippedFile.Close()

		extractedFilePath := filepath.Join(targetDir, file.Name)
		if file.FileInfo().IsDir() {
			log.Println("Directory Created:", extractedFilePath)
			if os.Stat(extractedFilePath); os.IsNotExist(err) {
				// create the folder if it does not exist
				err := os.MkdirAll(extractedFilePath, 0755)
				if err != nil {
					return nil, fmt.Errorf("directory extraction failed: %s", err)
				}
			} else {
				// chmod to 755
				if err := os.Chmod(extractedFilePath, 0755); err != nil {
					return nil, fmt.Errorf("failed to chmod %s to 0755: %s", extractedFilePath, err)
				}
			}
		} else {
			log.Println("File extracted:", file.Name)
			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				return nil, fmt.Errorf("file extraction failed: %s", err)
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				return nil, fmt.Errorf("file data extraction failed: %s", err)
			}
		}
	}

	if viper.GetBool(config.ENABLE_PACKAGE_SETUP_HOOK_KEY) {
		// for now ignore the setup error
		pkg.RunSetup(targetDir)
	}

	return pkg.Manifest, nil
}

func (pkg *zipPackage) VerifyChecksum(checksum string) (bool, error) {
	sha, err := packageChecksum(pkg.ZipFile)
	if err != nil {
		return false, fmt.Errorf("failed to calculate checksum of package %s@%s", pkg.Name(), pkg.Version())
	}
	remoteChecksum := fmt.Sprintf("%x", sha)
	if remoteChecksum != checksum {
		return false, fmt.Errorf("package %s@%s has a wrong checksum, expected %s, but get %s", pkg.Name(), pkg.Version(), checksum, remoteChecksum)
	}
	return true, nil
}

func (pkg *defaultPackage) VerifySignature(signature string) (bool, error) {
	// TODO: implement the signature verification
	return true, nil
}

func packageChecksum(pkgFile string) ([]byte, error) {
	f, err := os.Open(pkgFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
