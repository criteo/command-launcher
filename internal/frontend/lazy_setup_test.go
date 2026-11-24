package frontend

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/criteo/command-launcher/internal/backend"
	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/criteo/command-launcher/internal/pkg"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// makePackageSource writes a package manifest.mf under dir/pkgName with an
// executable command (named cmdName) and, when withHook is set, a __setup__ hook
// that appends a line to <pkgDir>/setup.count.
func makePackageSource(t *testing.T, name string, dir string, pkgName string, cmdName string, withHook bool) *backend.PackageSource {
	t.Helper()
	pkgDir := filepath.Join(dir, pkgName)
	err := os.MkdirAll(pkgDir, 0755)
	assert.Nil(t, err)

	hook := ""
	if withHook {
		hook = `{"name": "__setup__", "type": "system", "executable": "sh", "args": ["-c", "echo run >> #CACHE#/setup.count"]},`
	}
	manifest := `{
  "pkgName": "` + pkgName + `",
  "version": "1.0.0",
  "cmds": [
    ` + hook + `
    {"name": "` + cmdName + `", "type": "executable", "group": "", "short": "test", "executable": "echo"}
  ]
}`
	err = os.WriteFile(filepath.Join(pkgDir, "manifest.mf"), []byte(manifest), 0644)
	assert.Nil(t, err)

	return &backend.PackageSource{
		Name:       name,
		RepoDir:    dir,
		SyncPolicy: backend.SYNC_POLICY_NEVER,
		IsManaged:  false,
	}
}

// newLazySetupFrontend builds a real backend from two on-disk packages: "run" belongs
// to a package with a __setup__ hook, "nohook" to a package without one.
func newLazySetupFrontend(t *testing.T) (*defaultFrontend, command.Command, command.Command) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the __setup__ fixture requires sh")
	}

	homeDir := t.TempDir()
	dropinDir := t.TempDir()
	defaultDir := t.TempDir()

	dropinSrc := makePackageSource(t, "dropin", dropinDir, "nohook-pkg", "nohook", false)
	defaultSrc := makePackageSource(t, "default", defaultDir, "lazy-pkg", "run", true)

	be, err := backend.NewDefaultBackend(homeDir, nil, dropinSrc, defaultSrc)
	assert.Nil(t, err)

	fe := &defaultFrontend{backend: be, appCtx: context.InitContext("cdt", "1.0.0", "0")}

	runCmd, err := be.FindCommand("", "run")
	assert.Nil(t, err)
	nohookCmd, err := be.FindCommand("", "nohook")
	assert.Nil(t, err)

	return fe, runCmd, nohookCmd
}

func setLazySetupConfig(t *testing.T, enabled bool) {
	t.Helper()
	previous := viper.GetBool(config.ENABLE_LAZY_SETUP_KEY)
	viper.Set(config.ENABLE_LAZY_SETUP_KEY, enabled)
	t.Cleanup(func() { viper.Set(config.ENABLE_LAZY_SETUP_KEY, previous) })
}

func TestLazySetup_RunsOnceBeforeCommand(t *testing.T) {
	fe, iCmd, _ := newLazySetupFrontend(t)
	setLazySetupConfig(t, true)

	pkgDir := iCmd.PackageDir()
	countFile := filepath.Join(pkgDir, "setup.count")

	// setup hasn't run yet
	_, err := os.Stat(countFile)
	assert.True(t, os.IsNotExist(err))
	assert.False(t, pkg.IsSetupDone(pkgDir))

	exitCode, err := fe.executeCommand("", "run", []string{}, []string{}, []string{})
	assert.Nil(t, err)
	assert.Equal(t, 0, exitCode)

	data, err := os.ReadFile(countFile)
	assert.Nil(t, err)
	assert.Equal(t, "run\n", string(data))
	assert.True(t, pkg.IsSetupDone(pkgDir))

	// second invocation must not re-run setup
	exitCode, err = fe.executeCommand("", "run", []string{}, []string{}, []string{})
	assert.Nil(t, err)
	assert.Equal(t, 0, exitCode)

	data, err = os.ReadFile(countFile)
	assert.Nil(t, err)
	assert.Equal(t, "run\n", string(data))
}

func TestLazySetup_PackageWithoutHook(t *testing.T) {
	fe, _, nohookCmd := newLazySetupFrontend(t)
	setLazySetupConfig(t, true)

	// the command runs and no marker is written for a package without __setup__
	exitCode, err := fe.executeCommand("", "nohook", []string{}, []string{}, []string{})
	assert.Nil(t, err)
	assert.Equal(t, 0, exitCode)
	assert.False(t, pkg.IsSetupDone(nohookCmd.PackageDir()))
}

func TestLazySetup_DisabledByConfig(t *testing.T) {
	fe, iCmd, _ := newLazySetupFrontend(t)
	setLazySetupConfig(t, false)

	pkgDir := iCmd.PackageDir()
	countFile := filepath.Join(pkgDir, "setup.count")

	// the command still runs, but the setup hook is not triggered and no marker is written
	exitCode, err := fe.executeCommand("", "run", []string{}, []string{}, []string{})
	assert.Nil(t, err)
	assert.Equal(t, 0, exitCode)

	_, err = os.Stat(countFile)
	assert.True(t, os.IsNotExist(err))
	assert.False(t, pkg.IsSetupDone(pkgDir))
}
