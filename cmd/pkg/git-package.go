package pkg

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/criteo/command-launcher/internal/command"
)

type gitPackage struct {
	folderPackage
}

func CreateGitRepo(urlAsString string) (command.Package, error) {
	if _, err := url.Parse(urlAsString); err != nil {
		return nil, fmt.Errorf("invalid url or pathname: %s (%v)", urlAsString, err)
	}

	cloneDir, err := os.MkdirTemp("", "git-package-*")
	if err != nil {
		return nil, fmt.Errorf("cannot create the folder to clone the git repo: %v", err)
	}

	mf, err := cloneRepo(urlAsString, cloneDir)
	if err != nil {
		return nil, fmt.Errorf("git command has failed: %v", err)
	}

	pkg := gitPackage{
		folderPackage: folderPackage{
			defaultPackage: defaultPackage{
				Manifest: mf,
			},
			sourceDir: cloneDir,
		},
	}

	return &pkg, fmt.Errorf("not implemented yet")
}

func cloneRepo(gitUrl string, targetDir string) (command.PackageManifest, error) {
	ctx := exec.Command("git", "clone", gitUrl)
	ctx.Dir = targetDir
	ctx.Stdout = os.Stdout
	ctx.Stderr = os.Stderr
	ctx.Stdin = os.Stdin

	if err := ctx.Run(); err != nil {
		return nil, fmt.Errorf("git command has failed: %v", err)
	}

	repoDir := strings.ReplaceAll(filepath.Base(gitUrl), filepath.Ext(gitUrl), "")
	manifestFile, err := os.Open(filepath.Join(targetDir, repoDir, "manifest.mf"))
	if err != nil {
		return nil, err
	}
	defer manifestFile.Close()

	return ReadManifest(manifestFile)
}
