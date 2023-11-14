package server

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/frontend"
)

func (server *Server) CommandHandler(w http.ResponseWriter, r *http.Request) {
	fullName := strings.TrimPrefix(r.URL.Path, "/command/")
	backend := *(server.backend)
	cmd, err := backend.FindCommandByFullName(fullName)

	isRenamed := false
	alias := fmt.Sprintf("%s ", cmd.Group())
	renamedCmds := backend.AllRenamedCommands()

	if cmd.Group() != "" {
		// check if the group name is renamed
		fullName := fmt.Sprintf("%s@@%s@%s", cmd.Group(), cmd.PackageName(), cmd.RepositoryID())
		if renamedCmd, ok := renamedCmds[fullName]; ok {
			alias = fmt.Sprintf("%s ", renamedCmd)
			isRenamed = true
		}
	}
	if renamedCmd, ok := renamedCmds[cmd.FullName()]; ok {
		isRenamed = true
		alias = fmt.Sprintf("%s%s", alias, renamedCmd)
	} else {
		alias = fmt.Sprintf("%s%s", alias, cmd.Name())
	}

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	homeDir, _ := os.UserHomeDir()

	tmpl_text, _ := templates.ReadFile("templates/command.html")
	tmpl, _ := template.New("command").Parse(string(tmpl_text))
	command := Command{
		FullName:       cmd.FullName(),
		Group:          cmd.Group(),
		Name:           cmd.Name(),
		Package:        cmd.PackageName(),
		Registry:       cmd.RepositoryID(),
		Short:          cmd.ShortDescription(),
		Long:           cmd.LongDescription(),
		Examples:       cmd.Examples(),
		Flags:          getCmdFlags(cmd),
		DefaultWorkDir: homeDir + "/.cdt",
		HasAlias:       isRenamed,
		Alias:          alias,
	}

	tmpl.Execute(w, command)
}

func getCmdFlags(cmd command.Command) []command.Flag {
	flags := cmd.Flags()

	legacyFlags := cmd.RequiredFlags()
	for _, legacyFlag := range legacyFlags {
		flagName, flagShort, flagDesc, flagType, defaultValue := frontend.ParseFlagDefinition(legacyFlag)
		flags = append(flags, command.Flag{
			FlagName:        strings.TrimLeft(flagName, "--"),
			FlagType:        flagType,
			FlagShortName:   flagShort,
			FlagDescription: flagDesc,
			FlagDefault:     defaultValue,
		})
	}
	return flags
}
