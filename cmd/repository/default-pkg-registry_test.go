package repository

import (
	"fmt"
	"testing"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/stretchr/testify/assert"
)

func generateTestRegistryFile(reg Registry, numOfPkgs int, numOfCmds int) error {
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
				CmdCheckFlags:       false,
				PkgDir:              "",
			}
			pkg.PkgCommands = append(pkg.PkgCommands, &cmd)
		}
		err := reg.Add(&pkg)
		if err != nil {
			return err
		}
	}

	return nil
}

func Test_defaultRegistry_Add(t *testing.T) {
	nbOfPkgs := 10
	nbOfCmds := 5

	reg, err := newDefaultRegistry()
	assert.Nil(t, err)
	err = generateTestRegistryFile(reg, nbOfPkgs, nbOfCmds)
	assert.Nil(t, err)

	pkgs := reg.AllPackages()
	assert.Equal(t, nbOfPkgs, len(pkgs), fmt.Sprintf("there must be %d packages", nbOfPkgs))

	exeCmds := reg.ExecutableCommands()
	assert.Equal(t, nbOfPkgs*nbOfCmds, len(exeCmds), fmt.Sprintf("there must be %d executable cmds", nbOfPkgs*nbOfCmds))

	groupCmds := reg.GroupCommands()
	assert.Equal(t, 0, len(groupCmds), "there should be no group cmds")
}

func Test_defaultRegistry_Remove(t *testing.T) {
	nbOfPkgs := 2
	nbOfCmds := 3

	reg, err := newDefaultRegistry()
	assert.Nil(t, err)
	err = generateTestRegistryFile(reg, nbOfPkgs, nbOfCmds)
	assert.Nil(t, err)

	err = reg.Remove("test-0")
	assert.Nil(t, err)
	nbOfPkgs -= 1

	pkgs := reg.AllPackages()
	assert.Equal(t, nbOfPkgs, len(pkgs), fmt.Sprintf("there must be %d packages", nbOfPkgs))
	exeCmds := reg.ExecutableCommands()
	assert.Equal(t, nbOfPkgs*nbOfCmds, len(exeCmds), fmt.Sprintf("there must be %d executable cmds", nbOfPkgs*nbOfCmds))

	err = reg.Remove("test-1")
	assert.Nil(t, err)
	nbOfPkgs -= 1

	pkgs = reg.AllPackages()
	assert.Equal(t, nbOfPkgs, len(pkgs), fmt.Sprintf("there must be %d packages", nbOfPkgs))
	exeCmds = reg.ExecutableCommands()
	assert.Equal(t, nbOfPkgs*nbOfCmds, len(exeCmds), fmt.Sprintf("there must be %d executable cmds", nbOfPkgs*nbOfCmds))
}

func Test_defaultRegistry_Update(t *testing.T) {
	nbOfPkgs := 5
	nbOfCmds := 3

	reg, err := newDefaultRegistry()
	assert.Nil(t, err)
	err = generateTestRegistryFile(reg, nbOfPkgs, nbOfCmds)
	assert.Nil(t, err)

	pkg := defaultRegistryEntry{
		PkgName:     fmt.Sprintf("test-%d", nbOfPkgs-2),
		PkgVersion:  "1.0.0",
		PkgCommands: []*command.DefaultCommand{},
	}

	err = reg.Update(&pkg)
	assert.Nil(t, err)

	pkgs := reg.AllPackages()
	assert.Equal(t, nbOfPkgs, len(pkgs), fmt.Sprintf("there must be %d packages", nbOfPkgs))
	exeCmds := reg.ExecutableCommands() // the new package has no commands
	assert.Equal(t, (nbOfPkgs-1)*nbOfCmds, len(exeCmds), fmt.Sprintf("there must be %d executable cmds", (nbOfPkgs-1)*nbOfCmds))
}

func Test_defaultRegistry_Query(t *testing.T) {
	nbOfPkgs := 10
	nbOfCmds := 5

	reg, err := newDefaultRegistry()
	assert.Nil(t, err)
	err = generateTestRegistryFile(reg, nbOfPkgs, nbOfCmds)
	assert.Nil(t, err)

	pkg, err := reg.Package("test-0")
	assert.Nil(t, err)
	assert.NotNil(t, pkg)

	cmds := pkg.Commands()
	assert.Equal(t, nbOfCmds, len(cmds), fmt.Sprintf("there must be %d executable cmds", nbOfCmds))

	cmd, err := reg.Command("test-group", "test-1-2")
	assert.Nil(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, "executable", cmd.Type(), "The type must be executable")
}
