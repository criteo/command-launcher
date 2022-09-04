package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/stretchr/testify/assert"
)

func generateTestRegistryFile(path string, numOfPkgs int, numOfCmds int) (*defaultRegistry, error) {
	reg, err := LoadRegistry(path)
	if err != nil {
		return nil, err
	}

	for i := 0; i < numOfPkgs; i++ {
		pkg := defaultRegistryEntry{
			PkgName:     fmt.Sprintf("test-%d", i),
			PkgVersion:  "1.0.0",
			PkgCommands: []*command.DefaultCommand{},
		}

		for j := 0; j < numOfCmds; j++ {
			cmd := command.DefaultCommand{
				CmdName:             fmt.Sprintf("test-%d-%d", i, j),
				CmdType:             "executable",
				CmdGroup:            "test-group",
				CmdCategory:         "",
				CmdShortDescription: "Short Description",
				CmdLongDescription:  "Long Description",
				CmdExecutable:       "#CACHE#/#OS#/#ARCH#/test#EXT#",
				CmdArguments:        []string{"option1", "option2"},
				CmdDocFile:          "#CACHE#/doc/index.md",
				CmdDocLink:          "https://dummy/doc/",
				CmdValidArgs:        []string{"arg1", "arg2", "arg3"},
				CmdRequiredFlags:    []string{"moab", "moab-id"},
				PkgDir:              "",
			}
			pkg.PkgCommands = append(pkg.PkgCommands, &cmd)
		}
		err = reg.Add(pkg)
		if err != nil {
			return nil, err
		}
	}

	err = reg.Store(path)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

func TestStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test-reg.json")
	reg, err := generateTestRegistryFile(path, 1, 1)
	assert.Nil(t, err)

	exeCmds := reg.ExecutableCommands()
	assert.Equal(t, 1, len(exeCmds), "there must be 1 executable cmds")

	groupCmds := reg.GroupCommands()
	assert.Equal(t, 0, len(groupCmds), "there should be no group cmds")

	err = reg.Store(path)
	assert.Nil(t, err)

	_, err = os.Stat(path)
	assert.Nil(t, err)
	assert.False(t, os.IsNotExist(err))
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test-reg.json")
	_, err := generateTestRegistryFile(path, 1, 1)
	assert.Nil(t, err)

	reg, err := LoadRegistry(path)
	assert.Nil(t, err)

	exeCmds := reg.ExecutableCommands()
	assert.Equal(t, 1, len(exeCmds), "there should be 1 executable cmds")
}

func TestAddRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test-reg.json")
	reg, err := generateTestRegistryFile(path, 1, 1)
	assert.Nil(t, err)

	exeCmds := reg.ExecutableCommands()
	assert.Equal(t, 1, len(exeCmds), "there should be 1 executable cmd")

	fmt.Println(exeCmds[0])

	c, err := reg.Command("test-group", "test-0-0")
	assert.Nil(t, err)
	assert.Equal(t, "test-0-0", c.Name())
	assert.Equal(t, "Short Description", c.ShortDescription())

	err = reg.Remove("test-0")
	assert.Nil(t, err)

	exeCmds = reg.ExecutableCommands()
	assert.Equal(t, 0, len(exeCmds))
}

func BenchmarkLoadLargeRegistry(t *testing.B) {
	numOfPkg := 1000
	numOfCmd := 10

	path := filepath.Join(t.TempDir(), "test-reg.json")
	reg, err := generateTestRegistryFile(path, numOfPkg, numOfCmd)

	assert.Nil(t, err)

	exeCmds := reg.ExecutableCommands()
	assert.Equal(t, numOfCmd*numOfPkg, len(exeCmds), "there should be right number of executable cmd")

	start := time.Now()
	loadedReg, err := LoadRegistry(path)
	elapsed := time.Since(start)
	fmt.Println(elapsed)
	assert.Nil(t, err)
	loadedExeCmds := loadedReg.ExecutableCommands()
	assert.Equal(t, len(exeCmds), len(loadedExeCmds), "should have same number of executable cmd")

	// assert.Fail(t, "")
	// registry load time must less than 0.1 second (100 ms)
	assert.True(t, elapsed.Seconds() < 0.1)
}
