package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/criteo/command-launcher/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestLoadIndex(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "remote-test")
	err := os.Mkdir(basePath, 0755)
	assert.Nil(t, err)

	indexPath := filepath.Join(basePath, "index.json")
	err = helper.CopyLocalFile("assets/remote/basic-index.json", indexPath, false)
	assert.Nil(t, err)

	err = helper.CopyLocalFile("assets/ls-0.0.2.pkg", filepath.Join(basePath, "ls-0.0.2.pkg"), false)
	assert.Nil(t, err)

	remoteRepo := CreateRemoteRepository(fmt.Sprintf("file://%s", basePath))

	err = remoteRepo.Fetch()
	assert.Nil(t, err)

	pkgs, err := remoteRepo.PackageNames()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(pkgs))

	allPkgs, err := remoteRepo.All()
	assert.Nil(t, err)
	assert.Equal(t, 6, len(allPkgs))

	pkgNames, err := remoteRepo.PackageNames()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(pkgNames))
	ls := false
	hotfix := false
	env := false
	for _, pkgName := range pkgNames {
		if pkgName == "ls" {
			ls = true
			continue
		}
		if pkgName == "env" {
			env = true
			continue
		}
		if pkgName == "hotfix" {
			hotfix = true
			continue
		}
	}
	assert.True(t, ls)
	assert.True(t, hotfix)
	assert.True(t, env)

	versions, err := remoteRepo.Versions("hotfix")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(versions))
	assert.Equal(t, "1.0.0-43596", versions[0])
	assert.Equal(t, "1.0.0-43603", versions[1])
	assert.Equal(t, "1.0.0-43736", versions[2])

	version, err := remoteRepo.LatestVersion("ls")
	assert.Nil(t, err)
	assert.Equal(t, "0.0.2", version)

	version, err = remoteRepo.LatestVersion("hotfix")
	assert.Nil(t, err)
	assert.Equal(t, "1.0.0-43736", version)

	version, err = remoteRepo.LatestVersion("env")
	assert.Nil(t, err)
	assert.Equal(t, "1.0.0-43736", version)

	// query latest version for a particular partition
	version, err = remoteRepo.QueryLatestVersion("hotfix", func(pkgInfo *PackageInfo) bool {
		return pkgInfo.StartPartition <= 3 && pkgInfo.EndPartition >= 3
	})
	assert.Nil(t, err)
	assert.Equal(t, "1.0.0-43603", version)

	version, err = remoteRepo.QueryLatestVersion("hotfix", func(pkgInfo *PackageInfo) bool {
		return pkgInfo.StartPartition <= 7 && pkgInfo.EndPartition >= 7
	})
	assert.Nil(t, err)
	assert.Equal(t, "1.0.0-43736", version)

	// query experimental partition
	version, err = remoteRepo.QueryLatestVersion("env", func(pkgInfo *PackageInfo) bool {
		return pkgInfo.StartPartition >= 20
	})
	assert.Nil(t, err)
	assert.Equal(t, "1.0.0-43736", version)

	pkg, err := remoteRepo.Package("ls", "0.0.2")
	assert.Nil(t, err)
	assert.Equal(t, "ls", pkg.Name())
	assert.Equal(t, "0.0.2", pkg.Version())
	assert.Equal(t, 1, len(pkg.Commands()))
}
