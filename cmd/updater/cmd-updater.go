package updater

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/criteo/command-launcher/cmd/remote"
	"github.com/criteo/command-launcher/cmd/repository"
	"github.com/criteo/command-launcher/cmd/user"
	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/helper"

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
			console.Highlight("- %s command '%s' from version %s to version %s ...\n", op, pkgName, localPkg.Version(), remoteVersion)
			pkg, err := remoteRepo.Package(pkgName, remoteVersion)
			if err != nil {
				errPool = append(errPool, err)
				fmt.Printf("Cannot get the package of the command %s: %v\n", pkgName, err)
				continue
			}
			if err = repo.Update(pkg); err != nil {
				errPool = append(errPool, err)
				fmt.Printf("Cannot update the command %s: %v\n", pkgName, err)
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
					continue
				}
				if err = repo.Install(pkg); err != nil {
					errPool = append(errPool, err)
					fmt.Printf("Cannot install the package %s: %v\n", pkgName, err)
				}
			} else {
				errPool = append(errPool,
					fmt.Errorf("Package %s already exists in your local registry, you probably have a corrupted local registry", pkgName))
			}
		}
	}

	if len(errPool) == 0 {
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
			if lockedPkgs, err := u.loadLockedPackages(u.PackageLockFile); err == nil && len(lockedPkgs) > 0 {
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
func (u *CmdUpdater) loadLockedPackages(lockFile string) (map[string]string, error) {
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
