package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test-reg.json")
	reg, err := newJsonRegistry(path)
	assert.Nil(t, err)
	err = generateTestRegistry(reg, 1, 1)
	assert.Nil(t, err)

	exeCmds := reg.ExecutableCommands()
	assert.Equal(t, 1, len(exeCmds), "there must be 1 executable cmds")

	groupCmds := reg.GroupCommands()
	assert.Equal(t, 0, len(groupCmds), "there should be no group cmds")

	_, err = os.Stat(path)
	assert.Nil(t, err)
	assert.False(t, os.IsNotExist(err))
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test-reg.json")
	reg, err := newJsonRegistry(path)
	assert.Nil(t, err)
	err = generateTestRegistry(reg, 1, 1)
	assert.Nil(t, err)

	exeCmds := reg.ExecutableCommands()
	assert.Equal(t, 1, len(exeCmds), "there should be 1 executable cmds")
}

func TestAddRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test-reg.json")
	reg, err := newJsonRegistry(path)
	assert.Nil(t, err)
	err = generateTestRegistry(reg, 1, 1)
	assert.Nil(t, err)

	exeCmds := reg.ExecutableCommands()
	assert.Equal(t, 1, len(exeCmds), "there should be 1 executable cmd")

	fmt.Println(exeCmds[0])

	c, err := reg.Command("test-group", "test-0-0")
	assert.Nil(t, err)
	assert.Equal(t, "test-0-0", c.Name())
	assert.Equal(t, "Short Description", c.ShortDescription())

	err = reg.Remove("test-0", "")
	assert.Nil(t, err)

	exeCmds = reg.ExecutableCommands()
	assert.Equal(t, 0, len(exeCmds))
}

func BenchmarkLoadLargeRegistry(t *testing.B) {
	numOfPkg := 1000
	numOfCmd := 10

	path := filepath.Join(t.TempDir(), "test-reg.json")
	reg, err := newJsonRegistry(path)
	assert.Nil(t, err)
	err = generateTestRegistry(reg, numOfPkg, numOfCmd)
	assert.Nil(t, err)

	exeCmds := reg.ExecutableCommands()
	assert.Equal(t, numOfCmd*numOfPkg, len(exeCmds), "there should be right number of executable cmd")

	start := time.Now()
	loadedReg, err := newJsonRegistry(path)
	elapsed := time.Since(start)
	fmt.Println(elapsed)
	assert.Nil(t, err)
	loadedExeCmds := loadedReg.ExecutableCommands()
	assert.Equal(t, len(exeCmds), len(loadedExeCmds), "should have same number of executable cmd")
}
