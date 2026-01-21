package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/helper"
	"github.com/criteo/command-launcher/internal/remote"
	"github.com/criteo/command-launcher/internal/repository"
	"github.com/criteo/command-launcher/internal/user"

	log "github.com/sirupsen/logrus"
)

type CmdUpdater struct {
	cmdUpdateChan <-chan bool

	initRemoteRepoOnce sync.Once
	remoteRepo         remote.RemoteRepository
	initRemoteRepoErr  error

	toBeDeleted   map[string]string
	toBeUpdated   map[string]string
	toBeInstalled map[string]string

	CmdRepositoryBaseUrl string
	LocalRepo            repository.PackageRepository
	User                 user.User
	Timeout              time.Duration
	EnableCI             bool
	PackageLockFile      string
	IgnoreUpdatePause    bool
	VerifyChecksum       bool
	VerifySignature      bool
	SyncPolicy           string
}

func (u *CmdUpdater) CheckUpdateAsync() {
	ch := make(chan bool, 1)
	u.cmdUpdateChan = ch
	go func() {
		select {
		case value := <-u.checkUpdateCommands():
			ch <- value
		case <-time.After(u.Timeout):
			ch <- false
		}
	}()
}

func (u *CmdUpdater) Update() error {
	canBeUpdated := <-u.cmdUpdateChan
	if !canBeUpdated {
		return nil
	}

	errPool := []error{}

	// check if we are following the syncPolicy
	// TODO: for now we check the sync policy to block update during the update phase,
	// This is no optimal, as we still check remote repository in check update async.
	// We should move the sync policy check to the check update async, which will save
	// time to check the remote repo in order to make the check done in the timeout period.
	if err := u.reachSyncSchedule(); err != nil {
		log.Info(err.Error())
		return err
	}

	remoteRepo, err := u.getRemoteRepository()
	if err != nil {
		// TODO: handle error here
		return err
	}

	fmt.Println("\n-----------------------------------")
	fmt.Println("Some commands require update, please wait...")
	repo := u.LocalRepo

	// first delete deprecated packages
	if u.toBeDeleted != nil && len(u.toBeDeleted) > 0 {
		for pkg := range u.toBeDeleted {
			console.Highlight("- remove deprecated package '%s', it will not be available from now on\n", pkg)
			if err = repo.Uninstall(pkg); err != nil {
				errPool = append(errPool, err)
				fmt.Printf("Cannot uninstall the package %s: %v\n", pkg, err)
			}
		}
	}

	// update existing pacakges
	if u.toBeUpdated != nil && len(u.toBeUpdated) > 0 {
		for pkgName, remoteVersion := range u.toBeUpdated {
			localPkg, err := u.LocalRepo.Package(pkgName)
			if err != nil {
				errPool = append(errPool, err)
				continue
			}
			op := "upgrade"
			if remote.IsVersionSmaller(remoteVersion, localPkg.Version()) {
				op = "downgrade"
			}
			console.Highlight("- %s package '%s' from version %s to version %s ...\n", op, pkgName, localPkg.Version(), remoteVersion)
			pkg, err := remoteRepo.Package(pkgName, remoteVersion)
			if err != nil {
				errPool = append(errPool, err)
				fmt.Printf("Cannot get the package %s: %v\n", pkgName, err)
				u.pausePackageOnFailure(pkgName)
				continue
			}
			if ok, err := remoteRepo.Verify(pkg, u.VerifyChecksum, u.VerifySignature); !ok || err != nil {
				errPool = append(errPool, err)
				fmt.Printf("Failed to verify package %s, skip it: %v\n", pkgName, err)
				u.pausePackageOnFailure(pkgName)
				continue
			}
			if err = repo.Update(pkg); err != nil {
				errPool = append(errPool, err)
				fmt.Printf("Cannot update the package %s: %v\n", pkgName, err)
				// Note: repo.Update() calls repo.Install() which already handles pausing on failure
			}
		}
	}

	// install new ones
	if u.toBeInstalled != nil && len(u.toBeInstalled) > 0 {
		for pkgName, remoteVersion := range u.toBeInstalled {
			_, err = repo.Package(pkgName)
			if err != nil { // only install package that doesn't exist locally
				console.Highlight("- install new package '%s'\n", pkgName)
				pkg, err := remoteRepo.Package(pkgName, remoteVersion)
				if err != nil {
					errPool = append(errPool, err)
					fmt.Printf("Cannot get the package %s: %v\n", pkgName, err)
					u.pausePackageOnFailure(pkgName)
					continue
				}
				if ok, err := remoteRepo.Verify(pkg, u.VerifyChecksum, u.VerifySignature); !ok || err != nil {
					errPool = append(errPool, err)
					fmt.Printf("Failed to verify package %s, skip it: %v\n", pkgName, err)
					u.pausePackageOnFailure(pkgName)
					continue
				}
				if err = repo.Install(pkg); err != nil {
					errPool = append(errPool, err)
					fmt.Printf("Cannot install the package %s: %v\n", pkgName, err)
					// Note: repo.Install() already handles pausing on failure
				}
			} else {
				errPool = append(errPool,
					fmt.Errorf("Package %s already exists in your local registry, you probably have a corrupted local registry", pkgName))
			}
		}
	}

	if len(errPool) == 0 {
		// update the sync timestamp
		err := u.UpdateSyncTimestamp()
		if err != nil {
			log.Error(err)
		}

		fmt.Println("Update done! Enjoy coding!")
		return nil
	} else {
		return errPool[0]
	}
}

func (u *CmdUpdater) checkUpdateCommands() <-chan bool {
	ch := make(chan bool, 1)
	canBeUpdated := false
	go func() {
		remoteRepo, err := u.getRemoteRepository()
		if err != nil {
			canBeUpdated = false
			ch <- canBeUpdated
			return
		}

		install := map[string]string{}
		update := map[string]string{}
		delete := map[string]string{}

		// find all available package for this user's partition
		availablePkgs := map[string]string{}
		if remotePkgNames, err := remoteRepo.PackageNames(); err == nil {
			for _, remotePkgName := range remotePkgNames {
				latest, err := remoteRepo.QueryLatestPackageInfo(remotePkgName, func(pkgInfo *remote.PackageInfo) bool {
					return u.User.InPartition(pkgInfo.StartPartition, pkgInfo.EndPartition)
				})
				if err != nil {
					continue
				}
				availablePkgs[latest.Name] = latest.Version
			}
		}

		if u.EnableCI {
			log.Infoln("CI mode enabled")
			if lockedPkgs, err := u.LoadLockedPackages(u.PackageLockFile); err == nil && len(lockedPkgs) > 0 {
				log.Infof("checking locked packages from %s ...", u.PackageLockFile)
				// check if the locked packages are in the remote registry
				for k, v := range lockedPkgs {
					log.Infof("package %s is locked to version %s", k, v)
					if _, ok := availablePkgs[k]; !ok {
						log.Infoln(fmt.Errorf("package %s@%s is not available on the remote registry", k, v))
						canBeUpdated = false
						ch <- canBeUpdated
						return
					}
					// TODO: check if the locked version exists
				}
				// now set available packages to the locked ones
				availablePkgs = lockedPkgs
			} else if err != nil {
				log.Errorln(err)
			} else {
				log.Infof("Empty lock file %s", u.PackageLockFile)
			}
		}

		// iterate local packages to find to be deleted and to be updated ones
		// delete : exist in local, but not in remote
		// update: exist both in local and remote, but different versions
		localPkgMap := map[string]string{}
		localPkgs := u.LocalRepo.InstalledPackages()
		for _, localPkg := range localPkgs {
			localPkgMap[localPkg.Name()] = localPkg.Version()
			if remoteVersion, exist := availablePkgs[localPkg.Name()]; exist {
				if !u.IgnoreUpdatePause {
					paused, err := u.LocalRepo.IsPackageUpdatePaused(localPkg.Name())
					if err != nil {
						log.Errorf("Cannot check if package %s is paused: %v", localPkg.Name(), err)
					}
					if paused {
						// skip paused packages
						continue
					}
				}
				if remoteVersion != localPkg.Version() {
					// to be updated
					update[localPkg.Name()] = remoteVersion
				}
			} else {
				// to be deleted
				delete[localPkg.Name()] = localPkg.Version()
			}
		}

		// iterate the available pacakge again to find to be newly installed ones
		// (exist in remote available pkgs, but not in local packages)
		for pkg, version := range availablePkgs {
			if _, exist := localPkgMap[pkg]; !exist {
				// Check if the new package is paused (e.g., from a previous failed installation)
				if !u.IgnoreUpdatePause {
					paused, err := u.LocalRepo.IsPackageUpdatePaused(pkg)
					if err != nil {
						log.Errorf("Cannot check if package %s is paused: %v", pkg, err)
					}
					if paused {
						// skip paused packages
						log.Infof("Skipping paused package %s", pkg)
						continue
					}
				}
				install[pkg] = version
			}
		}

		u.toBeDeleted = delete
		u.toBeUpdated = update
		u.toBeInstalled = install

		if len(u.toBeDeleted) > 0 || len(u.toBeUpdated) > 0 || len(u.toBeInstalled) > 0 {
			canBeUpdated = true
		} else {
			canBeUpdated = false
		}

		ch <- canBeUpdated
	}()

	return ch
}

// only fetch remote repository once in each updater instance
func (u *CmdUpdater) getRemoteRepository() (remote.RemoteRepository, error) {
	if u.CmdRepositoryBaseUrl == "" {
		return nil, fmt.Errorf("invalid remote repository url")
	}
	u.initRemoteRepoOnce.Do(func() {
		u.remoteRepo = remote.CreateRemoteRepository(u.CmdRepositoryBaseUrl)
		u.initRemoteRepoErr = u.remoteRepo.Fetch()
	})
	return u.remoteRepo, u.initRemoteRepoErr
}

// load the package lock file
func (u *CmdUpdater) LoadLockedPackages(lockFile string) (map[string]string, error) {
	lockedPkgs := map[string]string{}
	content, err := helper.LoadFile(lockFile)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(content, &lockedPkgs); err != nil {
		return nil, err
	}
	return lockedPkgs, nil
}

// check sync policy
func (u *CmdUpdater) reachSyncSchedule() error {
	// check if we are following the syncPolicy
	if u.SyncPolicy == "never" {
		return errors.New(fmt.Sprintf("Remote '%s': Sync policy is set to never, no update will be performed", u.LocalRepo.Name()))
	}
	if u.SyncPolicy == "always" {
		return nil
	}
	// now load the sync timestamp
	localRepoFolder, err := u.LocalRepo.RepositoryFolder()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path.Join(localRepoFolder, "sync.timestamp"))
	if err != nil {
		// error read the file, we assume the sync time is passed
		return nil
	}
	syncTime, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return err
	}

	// now check if we passed the sync time
	if time.Now().Before(syncTime) {
		return errors.New(fmt.Sprintf("Remote '%s': Not yet reach the sync time", u.LocalRepo.Name()))
	}

	return nil
}

func (u *CmdUpdater) UpdateSyncTimestamp() error {
	localRepoFolder, err := u.LocalRepo.RepositoryFolder()
	if err != nil {
		return err
	}

	var delay time.Duration = 24
	switch u.SyncPolicy {
	case "always":
		return errors.New(fmt.Sprintf("Remote '%s': Sync policy is set to always, no need to update the sync timestamp", u.LocalRepo.Name()))
	case "never":
		return errors.New(fmt.Sprintf("Remote '%s': Sync policy is set to never, no need to update the sync timestamp", u.LocalRepo.Name()))
	case "hourly":
		delay = 1
	case "daily":
		delay = 24
	case "weekly":
		delay = 24 * 7
	case "monthly":
		delay = 24 * 30
	}

	err = os.WriteFile(path.Join(localRepoFolder, "sync.timestamp"), []byte(time.Now().Add(time.Hour*delay).Format(time.RFC3339)), 0644)

	log.Infof("Remote '%s': Sync timestamp updated to %s", u.LocalRepo.Name(), time.Now().Add(time.Hour*delay).Format(time.RFC3339))
	return err
}

// pausePackageOnFailure pauses a package after an installation failure
func (u *CmdUpdater) pausePackageOnFailure(pkgName string) {
	if err := u.LocalRepo.PausePackageUpdate(pkgName); err != nil {
		console.Warn("Failed to pause update for package %s: %v", pkgName, err)
	} else {
		console.Reminder(
			"Package %s has been paused due to installation failure, explicitly run `update package` to retry installation.",
			pkgName,
		)
	}
}
