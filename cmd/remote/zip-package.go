package remote

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/criteo/command-launcher/internal/command"
	log "github.com/sirupsen/logrus"
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
			err := os.MkdirAll(extractedFilePath, file.Mode())
			if err != nil {
				return nil, fmt.Errorf("directory extraction failed: %s", err)
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

	return pkg.Manifest, nil
}
