package backend

import (
	"fmt"
	"strings"
	"time"

	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/remote"
	"github.com/criteo/command-launcher/internal/repository"
	"github.com/criteo/command-launcher/internal/updater"
	"github.com/criteo/command-launcher/internal/user"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

const (
	SYNC_POLICY_NEVER   = "never"
	SYNC_POLICY_ALWAYS  = "always"
	SYNC_POLICY_HOURLY  = "hourly"
	SYNC_POLICY_DAILY   = "daily"
	SYNC_POLICY_WEEKLY  = "weekly"
	SYNC_POLICY_MONTHLY = "monthly"
)

type PackageSource struct {
	RepoDir           string
	RemoteBaseURL     string
	RemoteRegistryURL string
	SyncPolicy        string
	EnableSync        bool
	IsManaged         bool

	Repo    repository.PackageRepository
	Failure error

	Updater *updater.CmdUpdater
}

func NewDropinSource(repoDir string) *PackageSource {
	return &PackageSource{
		RepoDir:           repoDir,
		RemoteBaseURL:     "",
		RemoteRegistryURL: "",
		EnableSync:        false,
		IsManaged:         false,
		SyncPolicy:        SYNC_POLICY_NEVER,
	}
}

func NewManagedSource(repoDir, remoteBaseURL string, enableSync bool, syncPolicy string) *PackageSource {
	return &PackageSource{
		RepoDir:           repoDir,
		RemoteBaseURL:     remoteBaseURL,
		RemoteRegistryURL: fmt.Sprintf("%s/index.json", remoteBaseURL),
		EnableSync:        enableSync,
		IsManaged:         true,
		SyncPolicy:        SYNC_POLICY_ALWAYS,
	}
}

func (src *PackageSource) InitUpdater(user *user.User, timeout time.Duration, enableCI bool, lockFile string, verifyChecksum bool, verifySignature bool) *updater.CmdUpdater {
	if !src.EnableSync {
		return nil
	}
	src.Updater = &updater.CmdUpdater{
		LocalRepo:            src.Repo,
		CmdRepositoryBaseUrl: src.RemoteBaseURL,
		User:                 *user,
		Timeout:              timeout,
		EnableCI:             enableCI,
		PackageLockFile:      lockFile,
		VerifyChecksum:       verifyChecksum,
		VerifySignature:      verifySignature,
	}
	return src.Updater
}

func (src PackageSource) IsInstalled() bool {
	if src.Repo == nil {
		// nothing to install
		return true
	}
	return len(src.Repo.InstalledCommands()) > 0
}

func (src *PackageSource) InitialInstallCommands(user *user.User) error {
	remote := remote.CreateRemoteRepository(src.RemoteBaseURL)
	errors := make([]string, 0)

	// check locked packages if ci is enabled
	lockedPackages := map[string]string{}
	if viper.GetBool(config.CI_ENABLED_KEY) {
		pkgs, err := src.Updater.LoadLockedPackages(viper.GetString(config.PACKAGE_LOCK_FILE_KEY))
		if err == nil {
			lockedPackages = pkgs
		}
	}

	if pkgs, err := remote.PackageNames(); err == nil {
		for _, pkgName := range pkgs {
			pkgVersion := "unspecified"
			if lockedVersion, ok := lockedPackages[pkgName]; ok {
				pkgVersion = lockedVersion
			} else {
				latest, err := remote.LatestPackageInfo(pkgName)
				if err != nil {
					log.Error(err)
					errors = append(errors, fmt.Sprintf("cannot get the latest version of the package %s: %v", latest.Name, err))
					continue
				}
				if !user.InPartition(latest.StartPartition, latest.EndPartition) {
					log.Infof("Skip installing package %s, user not in partition (%d %d)\n", latest.Name, latest.StartPartition, latest.EndPartition)
					continue
				}
				pkgVersion = latest.Version
			}

			pkg, err := remote.Package(pkgName, pkgVersion)
			if err != nil {
				log.Error(err)
				errors = append(errors, fmt.Sprintf("cannot get the package %s: %v", pkgName, err))
				continue
			}
			if ok, err := remote.Verify(pkg,
				viper.GetBool(config.VERIFY_PACKAGE_CHECKSUM_KEY),
				viper.GetBool(config.VERIFY_PACKAGE_SIGNATURE_KEY),
			); !ok || err != nil {
				log.Error(err)
				errors = append(errors, fmt.Sprintf("failed to verify package %s, skip it: %v", pkgName, err))
				continue
			}
			err = src.Repo.Install(pkg)
			if err != nil {
				errors = append(errors, fmt.Sprintf("cannot install the package %s: %v", pkgName, err))
				continue
			}
		}
	} else {
		errors = append(errors, fmt.Sprintf("cannot get remote packages: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("install failed for the following reasons: [%s]", strings.Join(errors, ", "))
	}

	return nil
}
