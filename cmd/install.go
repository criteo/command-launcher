package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/criteo/command-launcher/cmd/remote"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/criteo/command-launcher/internal/helper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type InstallFlags struct {
	gitUrl  string
	fileUrl string
	version string
}

var (
	installFlags = InstallFlags{}
)

func AddInstallCmd(rootCmd *cobra.Command, appCtx context.LauncherContext) {
	packageInstallCmd := &cobra.Command{
		Use:               "install [package_name]",
		Short:             "Install a package",
		Long:              "Install a package package from a git repo or from a zip file or from its name",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: packageNameValidatonFunc(true),
		Example: fmt.Sprintf(`
  %s package install --git https://example.com/my-repo.git`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			if installFlags.fileUrl != "" {
				return installZipFile(installFlags.fileUrl)
			}

			if installFlags.gitUrl != "" {
				return installGitRepo(installFlags.gitUrl)
			}

			name := args[0]
			return installRemotePackage(name, installFlags.version)
		},
	}
	packageInstallCmd.Flags().StringVar(&installFlags.fileUrl, "file", "", "URL or path of a package file")
	packageInstallCmd.Flags().StringVar(&installFlags.gitUrl, "git", "", "URL of a Git repo of package")
	packageInstallCmd.Flags().StringVar(&installFlags.version, "version", "", "version of ")
	packageInstallCmd.MarkFlagsMutuallyExclusive("git", "file", "version")

	rootCmd.AddCommand(packageInstallCmd)
}

func installRemotePackage(name string, version string) error {
	return nil
}

func installGitRepo(gitUrl string) error {
	_, err := url.Parse(gitUrl)
	if err != nil {
		return fmt.Errorf("invalid url or pathname: %v", err)
	}

	ctx := exec.Command("git", "clone", gitUrl)
	ctx.Dir = viper.GetString(config.DROPIN_FOLDER_KEY)
	ctx.Stdout = os.Stdout
	ctx.Stderr = os.Stderr
	ctx.Stdin = os.Stdin

	if err = ctx.Run(); err != nil {
		return fmt.Errorf("git command has failed: %v", err)
	}

	repoDir := strings.ReplaceAll(filepath.Base(gitUrl), filepath.Ext(gitUrl), "")
	path := filepath.Join(viper.GetString(config.DROPIN_FOLDER_KEY), repoDir)
	pkg, err := remote.CreateFolderPackage(path)
	if err != nil {
		os.RemoveAll(path)
		return fmt.Errorf("the git repo does not contain a manifest file: %v", err)
	}

	console.Success("Package %s installed in the dropin repository", pkg.Name())
	return nil
}

func installZipFile(fileUrl string) error {
	url, err := url.Parse(fileUrl)
	if err != nil {
		return fmt.Errorf("invalid url or pathname: %v", err)
	}

	var pathname string
	if url.IsAbs() {
		if url.Scheme == "file" {
			pathname = url.Path
		} else {
			tmpDir, err := os.MkdirTemp("", "package-download-*")
			if err != nil {
				return fmt.Errorf("cannot create temporary dir (%v)", err)
			}
			defer os.RemoveAll(tmpDir)

			pkgPathname := filepath.Join(tmpDir, filepath.Base(url.Path))
			if err := helper.DownloadFile(fileUrl, pkgPathname, true); err != nil {
				return fmt.Errorf("error downloading %s: %v", url, err)
			}

		}
	} else {
		pathname = url.Path
	}

	pkg, err := remote.CreateZipPackage(pathname)
	if err != nil {
		return fmt.Errorf("cannot create the package from the zip file: %v", err)
	}

	mf, err := pkg.InstallTo(viper.GetString(config.DROPIN_FOLDER_KEY))
	if err == nil {
		console.Success("Package %s installed in the dropin repository", mf.Name())
	}

	return err
}
