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
	"github.com/criteo/command-launcher/internal/console"
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

	var backupDir string
	// If target directory exists, move it to backup location
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		tmpDir, err := os.MkdirTemp("", "package-backup-*")
		defer os.RemoveAll(tmpDir)
		if err != nil {
			return nil, fmt.Errorf("cannot create temporary backup directory: %v", err)
		}
		backupDir = filepath.Join(tmpDir, pkg.Name())

		// Create backup directory and copy existing target directory contents
		// to avoid cross-filesystem rename issues during restoration
		if err := os.CopyFS(backupDir, os.DirFS(targetDir)); err != nil {
			return nil, fmt.Errorf("cannot backup existing package directory %s: %v", targetDir, err)
		}
		if err := os.RemoveAll(targetDir); err != nil {
			return nil, fmt.Errorf("cannot remove existing package directory %s: %v", targetDir, err)
		}
	}
	// Create target directory
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("cannot create target package directory %s: %v", targetDir, err)
	}

	// Cleanup function to handle backup restoration and cleanup
	var installSuccessful bool
	defer RestoreBackupOnFailure(backupDir, targetDir, &installSuccessful)

	for _, file := range zipReader.Reader.File {
		if err := extractZipEntry(targetDir, file); err != nil {
			return nil, err
		}
	}

	var err error
	if viper.GetBool(config.ENABLE_PACKAGE_SETUP_HOOK_KEY) {
		err = pkg.RunSetup(targetDir)
		if err != nil {
			os.RemoveAll(targetDir)
			return nil, err
		}
	}

	// Mark installation as successful
	installSuccessful = true

	return pkg.Manifest, nil
}

func RestoreBackupOnFailure(backupDir string, targetDir string, installSuccessful *bool) {
	if backupDir != "" && !*installSuccessful {
		// Restore backup if install failed
		err := os.RemoveAll(targetDir)
		if err != nil {
			console.Error("Failed to remove target directory %s: %v", targetDir, err)
			return
		}
		err = os.CopyFS(targetDir, os.DirFS(backupDir))
		if err != nil {
			console.Error("Failed to restore backup from %s to %s: %v", backupDir, targetDir, err)
			return
		}
		err = os.RemoveAll(backupDir)
		if err != nil {
			console.Warn("Failed to remove backup directory %s: %v", backupDir, err)
			return
		}
		console.Warn("Restored the previous version of the package %s from backup", filepath.Base(backupDir))
	}
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

func extractZipEntry(targetDir string, file *zip.File) error {
	zippedFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("installation failed: %s", err)
	}
	defer zippedFile.Close()

	extractedFilePath := filepath.Join(targetDir, file.Name)
	if file.FileInfo().IsDir() {
		log.Println("Directory Created:", extractedFilePath)
		err := os.MkdirAll(extractedFilePath, file.Mode())
		if err != nil {
			return fmt.Errorf("directory extraction failed: %s", err)
		}

		fileStats, err := os.Stat(extractedFilePath)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %s", extractedFilePath, err)
		}
		permissions := fileStats.Mode().Perm()
		if permissions != 0o755 {
			// chmod to 755
			if err := os.Chmod(extractedFilePath, 0755); err != nil {
				return fmt.Errorf("failed to chmod %s to 0755: %s", extractedFilePath, err)
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
			return fmt.Errorf("file extraction failed: %s", err)
		}
		defer outputFile.Close()

		_, err = io.Copy(outputFile, zippedFile)
		if err != nil {
			return fmt.Errorf("file data extraction failed: %s", err)
		}
	}

	return nil
}
