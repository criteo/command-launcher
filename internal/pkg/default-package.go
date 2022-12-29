package pkg

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/criteo/command-launcher/internal/command"
	"gopkg.in/yaml.v3"
)

type defaultPackageManifest struct {
	PkgName     string                    `json:"pkgName" yaml:"pkgName"`
	PkgVersion  string                    `json:"version" yaml:"version"`
	PkgCommands []*command.DefaultCommand `json:"cmds" yaml:"cmds"`
}

func (mf *defaultPackageManifest) Name() string {
	return mf.PkgName
}

func (mf *defaultPackageManifest) Version() string {
	return mf.PkgVersion
}

func (mf *defaultPackageManifest) Commands() []command.Command {
	cmds := make([]command.Command, 0)
	for _, cmd := range mf.PkgCommands {
		//newCmd := cmd
		cmds = append(cmds, cmd)
	}
	return cmds
}

type defaultPackage struct {
	Manifest command.PackageManifest
	// store the repository id, indicates which registry/repository that the package belongs to
	repositoryID string
}

func (pkg *defaultPackage) Name() string {
	return pkg.Manifest.Name()
}

func (pkg *defaultPackage) Version() string {
	return pkg.Manifest.Version()
}

func (pkg *defaultPackage) Commands() []command.Command {
	return pkg.Manifest.Commands()
}

func (pkg *defaultPackage) RepositoryID() string {
	return pkg.repositoryID
}

func ReadManifest(file fs.File) (command.PackageManifest, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot read the manifest file handle (%s)", err)
	}

	var payload = make([]byte, stat.Size())
	nb, err := file.Read(payload)
	if err != nil && err != io.EOF || nb != int(stat.Size()) {
		return nil, fmt.Errorf("cannot read the manifest (%s)", err)
	}

	var mf = defaultPackageManifest{}
	// YAML is super set of json, should work with JSON as well
	err = yaml.Unmarshal(payload, &mf)
	if err != nil {
		return nil, fmt.Errorf("cannot read the manifest content, it is neither a valid JSON nor YAML (%s)", err)
	}

	return &mf, nil
}

func copyFolder(srcFolder string, dstFolder string) error {
	info, err := os.Stat(srcFolder)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dstFolder, info.Mode()); err != nil {
		return err
	}

	files, err := os.ReadDir(srcFolder)
	if err != nil {
		return err
	}

	for _, fd := range files {
		src := filepath.Join(srcFolder, fd.Name())
		dst := filepath.Join(dstFolder, fd.Name())

		if fd.IsDir() {
			if err = copyFolder(src, dst); err != nil {
				return err
			}
		} else {
			if err = copyFile(src, dst); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("file extraction failed: %s", err)
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}
