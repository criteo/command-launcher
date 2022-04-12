package remote

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/helper"
	log "github.com/sirupsen/logrus"
)

var (
	ErrMsg_PackageNotFound = "package not found"
)

func IsPackageNotFound(err error) bool {
	return strings.HasPrefix(err.Error(), ErrMsg_PackageNotFound)
}

type cdtRemoteRepository struct {
	repoBaseUrl    string
	PackagesByName map[string]PackagesByVersion
}

func newCdtRemoteRepository(baseUrl string) *cdtRemoteRepository {
	return &cdtRemoteRepository{
		repoBaseUrl:    baseUrl,
		PackagesByName: make(map[string]PackagesByVersion),
	}
}

func (remote *cdtRemoteRepository) Fetch() error {
	return remote.load()
}

func (remote *cdtRemoteRepository) All() ([]PackageInfo, error) {
	packages := make([]PackageInfo, 0)
	if err := remote.load(); err != nil {
		return packages, err
	}
	for _, list := range remote.PackagesByName {
		packages = append(packages, list...)
	}
	return packages, nil
}

func (remote *cdtRemoteRepository) PackageNames() ([]string, error) {
	packages := make([]string, 0)
	if err := remote.load(); err != nil {
		return packages, err
	}
	for key := range remote.PackagesByName {
		packages = append(packages, key)
	}
	return packages, nil
}

func (remote *cdtRemoteRepository) Versions(pkgName string) ([]string, error) {
	results := make([]string, 0)
	if err := remote.load(); err != nil {
		return results, err
	}
	pkgInfos, exists := remote.PackagesByName[pkgName]
	if exists {
		for _, info := range pkgInfos {
			var version cdtVersion
			err := ParseVersion(info.Version, &version)
			if err == nil && info.Name == pkgName {
				results = append(results, version.String())
			}
		}
	}

	return results, nil
}

func (remote *cdtRemoteRepository) PackageInfosByCmdName(pkgName string) ([]PackageInfo, error) {
	if err := remote.load(); err != nil {
		return []PackageInfo{}, err
	}
	pkgInfos, exists := remote.PackagesByName[pkgName]
	if exists {
		return pkgInfos, nil
	}
	return make(PackagesByVersion, 0), nil
}

func (remote *cdtRemoteRepository) LatestVersion(pkgName string) (string, error) {
	return remote.QueryLatestVersion(pkgName, func(pkgInfo *PackageInfo) bool {
		return true
	})
}

func (remote *cdtRemoteRepository) QueryLatestVersion(pkgName string, filter PackageInfoFilterFunc) (string, error) {
	pkgInfo, err := remote.QueryLatestPackageInfo(pkgName, filter)
	if err != nil {
		return "", err
	}
	if pkgInfo == nil {
		return "", fmt.Errorf("%s: %s", ErrMsg_PackageNotFound, pkgName)
	}
	return pkgInfo.Version, nil
}

func (remote *cdtRemoteRepository) LatestPackageInfo(pkgName string) (*PackageInfo, error) {
	return remote.QueryLatestPackageInfo(pkgName, func(pkgInfo *PackageInfo) bool {
		return true
	})
}

func (remote *cdtRemoteRepository) QueryLatestPackageInfo(pkgName string, filter PackageInfoFilterFunc) (*PackageInfo, error) {
	pkgInfos, err := remote.PackageInfosByCmdName(pkgName)
	if err != nil {
		return nil, err
	}
	if len(pkgInfos) == 0 {
		return nil, fmt.Errorf("%s in remote repository: %s", ErrMsg_PackageNotFound, pkgName)
	}
	for i := len(pkgInfos) - 1; i >= 0; i-- {
		if filter(&pkgInfos[i]) {
			return &pkgInfos[i], nil
		}
	}
	return nil, fmt.Errorf("%s in remote repository: %s. The package exists, but no version match the query filter", ErrMsg_PackageNotFound, pkgName)
}

func (remote *cdtRemoteRepository) Package(pkgName string, pkgVersion string) (command.Package, error) {
	tmpDir, err := os.MkdirTemp("", "package-download-*")
	if err != nil {
		return nil, fmt.Errorf("cannot create temporary dir (%v)", err)
	}

	pkgPathname := filepath.Join(tmpDir, fmt.Sprintf("%s-%s.pkg", pkgName, pkgVersion))

	url := remote.url(pkgName, pkgVersion)

	if err := helper.DownloadFile(url, pkgPathname, true); err != nil {
		return nil, fmt.Errorf("error downloading %s: %v", url, err)
	}

	pkg, err := CreatePackage(pkgPathname)
	if err != nil {
		return nil, fmt.Errorf("invalid package %s: %v", url, err)
	}

	return pkg, nil
}

func (remote *cdtRemoteRepository) url(name string, version string) string {
	return fmt.Sprintf("%s/%s", remote.repoBaseUrl, remote.pkgFilename(name, version))
}

func (remote *cdtRemoteRepository) pkgFilename(name string, version string) string {
	return fmt.Sprintf("%s-%s.pkg", name, version)
}

func (remote *cdtRemoteRepository) load() error {
	if !remote.isLoaded() {
		body, err := helper.LoadFile(fmt.Sprintf("%s/index.json", remote.repoBaseUrl))
		if err != nil {
			log.Error("Cannot read packages index")
			return err
		}

		var entries []PackageInfo
		err = json.Unmarshal(body, &entries)
		if err != nil {
			log.Error("json parsing error")
			return err
		}

		for _, pkg := range entries {
			lst, exists := remote.PackagesByName[pkg.Name]
			if exists {
				lst = append(lst, pkg)
				remote.PackagesByName[pkg.Name] = lst
			} else {
				newLst := make(PackagesByVersion, 0)
				newLst = append(newLst, pkg)
				remote.PackagesByName[pkg.Name] = newLst
			}
		}

		// sort packages
		for _, v := range remote.PackagesByName {
			sort.Sort(v)
		}
	}

	return nil
}

func (remote *cdtRemoteRepository) isLoaded() bool {
	return len(remote.PackagesByName) > 0
}
