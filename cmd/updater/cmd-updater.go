package updater

import (
	"fmt"
	"sync"
	"time"

	"github.com/criteo/command-launcher/cmd/remote"
	"github.com/criteo/command-launcher/cmd/repository"
	"github.com/criteo/command-launcher/cmd/user"
	"github.com/criteo/command-launcher/internal/console"
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

func (u *CmdUpdater) Update() {
	canBeUpdated := <-u.cmdUpdateChan
	if !canBeUpdated {
		return
	}

	remoteRepo, err := u.getRemoteRepository()
	if err != nil {
		// TODO: handle error here
		return
	}

	fmt.Println("\n-----------------------------------")
	fmt.Println("Some commands require update, please wait...")
	repo := u.LocalRepo

	// first delete deprecated packages
	if u.toBeDeleted != nil && len(u.toBeDeleted) > 0 {
		for pkg := range u.toBeDeleted {
			console.Highlight("- remove deprecated package '%s', it will not be available from now on\n", pkg)
			if err = repo.Uninstall(pkg); err != nil {
				fmt.Printf("Cannot uninstall the package %s: %v\n", pkg, err)
			}
		}
	}

	// update existing pacakges
	if u.toBeUpdated != nil && len(u.toBeUpdated) > 0 {
		for pkgName, remoteVersion := range u.toBeUpdated {
			localPkg, err := u.LocalRepo.Package(pkgName)
			if err != nil {
				continue
			}
			op := "upgrade"
			if remote.IsVersionSmaller(remoteVersion, localPkg.Version()) {
				op = "downgrade"
			}
			console.Highlight("- %s command '%s' from version %s to version %s ...\n", op, pkgName, localPkg.Version(), remoteVersion)
			pkg, err := remoteRepo.Package(pkgName, remoteVersion)
			if err != nil {
				fmt.Printf("Cannot get the package of the command %s: %v\n", pkgName, err)
				continue
			}
			if err = repo.Update(pkg); err != nil {
				fmt.Printf("Cannot update the command %s: %v\n", pkgName, err)
			}
		}
	}

	// install new ones
	if u.toBeInstalled != nil && len(u.toBeInstalled) > 0 {
		for pkgName, remoteVersion := range u.toBeInstalled {
			_, err = repo.Package(pkgName)
			if err != nil {
				console.Highlight("- install new package '%s'\n", pkgName)
				pkg, err := remoteRepo.Package(pkgName, remoteVersion)
				if err != nil {
					fmt.Printf("Cannot get the package %s: %v\n", pkgName, err)
					continue
				}
				if err = repo.Install(pkg); err != nil {
					fmt.Printf("Cannot install the package %s: %v\n", pkgName, err)
				}
			}
		}
	}

	fmt.Println("Update done! Enjoy coding!")
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
