package pkg

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type defaultPackageManifest struct {
	PkgName     string                    `json:"pkgName" yaml:"pkgName"`
	PkgVersion  string                    `json:"version" yaml:"version"`
	PkgCommands []*command.DefaultCommand `json:"cmds" yaml:"cmds"`
	PkgOptions  command.PackageOptions    `json:"options" yaml:"options"`
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

func (mf *defaultPackageManifest) Options() command.PackageOptions {
	return mf.PkgOptions
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

func (pkg *defaultPackage) Options() command.PackageOptions {
	return pkg.Manifest.Options()
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

	if mf.Options().ForceSetupOnInstall && !HasSetupHook(&mf) {
		log.Warnf("package %s sets options.forceSetupOnInstall but has no %s hook, the option is ignored", mf.Name(), SETUP_HOOK_NAME)
	}

	return &mf, nil
}

func (pkg *defaultPackage) RunSetup(pkgDir string) error {
	return ExecSetupHookFromPackage(pkg, pkgDir)
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

// SETUP_HOOK_NAME is the name of the system command a package may define as its setup hook.
const SETUP_HOOK_NAME = "__setup__"

func isSetupHook(c command.Command) bool {
	return c.Name() == SETUP_HOOK_NAME && c.Type() == "system"
}

// shouldRunSetupOnInstall reports whether the __setup__ hook must run at install time:
// the package has a __setup__ hook, and the setup is enabled by the user (ENABLE_PACKAGE_SETUP_HOOK)
// or forced by the package manifest.
func shouldRunSetupOnInstall(pkg command.PackageManifest) bool {
	if !HasSetupHook(pkg) {
		return false
	}
	if viper.GetBool(config.ENABLE_PACKAGE_SETUP_HOOK_KEY) {
		return true
	}
	return pkg.Options().ForceSetupOnInstall
}

// HasSetupHook reports whether the package defines a __setup__ system command.
func HasSetupHook(pkg command.PackageManifest) bool {
	for _, c := range pkg.Commands() {
		if isSetupHook(c) {
			return true
		}
	}
	return false
}

func ExecSetupHookFromPackage(pkg command.PackageManifest, pkgDir string) error {
	for _, c := range pkg.Commands() {
		if isSetupHook(c) {
			if pkgDir != "" {
				c.SetPackageDir(pkgDir)
			}
			// Execute the __setup__ hook
			_, err := c.Execute(os.Environ())
			if err != nil {
				return fmt.Errorf("setup hook of package %s failed to execute: %v", pkg.Name(), err)
			}
			if err := MarkSetupDone(c.PackageDir(), pkg.Version()); err != nil {
				log.Warnf("failed to write setup marker for package %s: %v", pkg.Name(), err)
			}
			return nil
		}
	}
	log.Warnf("No setup hook defined for package %s", pkg.Name())
	if dir := effectivePackageDir(pkg, pkgDir); dir != "" {
		if err := MarkSetupDone(dir, pkg.Version()); err != nil {
			log.Warnf("failed to write setup marker for package %s: %v", pkg.Name(), err)
		}
	}
	return nil
}

// effectivePackageDir returns pkgDir when provided, otherwise falls back to the
// package directory already set on one of its commands (populated when the
// package was loaded/registered).
func effectivePackageDir(pkg command.PackageManifest, pkgDir string) string {
	if pkgDir != "" {
		return pkgDir
	}
	for _, c := range pkg.Commands() {
		if c.PackageDir() != "" {
			return c.PackageDir()
		}
	}
	return ""
}
