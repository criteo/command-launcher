package server

import (
	"fmt"
	"net/http"
	"os"
	"os/user"
	"strings"

	"github.com/criteo/command-launcher/internal/helper"
)

func (server *Server) ExecuteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	backend := *(server.backend)
	// get the command full name from the path
	fullName := strings.TrimPrefix(r.URL.Path, "/execute/")
	cmd, err := backend.FindCommandByFullName(fullName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	r.ParseForm()

	renamedCmds := backend.AllRenamedCommands()

	// build command arguments
	args := []string{}
	if cmd.Group() != "" {
		// check if the group name is renamed
		fullName := fmt.Sprintf("%s@@%s@%s", cmd.Group(), cmd.PackageName(), cmd.RepositoryID())
		if renamedCmd, ok := renamedCmds[fullName]; ok {
			args = append(args, renamedCmd)
		} else {
			args = append(args, cmd.Group())
		}
	}

	if renamed, ok := renamedCmds[cmd.FullName()]; ok {
		args = append(args, renamed)
	} else {
		args = append(args, cmd.Name())
	}

	for _, flag := range cmd.Flags() {
		if r.Form.Has(flag.Name()) {
			v := r.Form.Get(flag.Name())
			if v == "" {
				continue
			}

			if flag.Type() == "string" {
				args = append(args, fmt.Sprintf("--%s", flag.Name()))
				args = append(args, fmt.Sprintf("%s", v))
			}
			if flag.Type() == "bool" && v == "on" {
				args = append(args, fmt.Sprintf("--%s", flag.Name()))
			}
		}
	}

	if r.Form.Has("__args__") {
		v := r.Form.Get("__args__")
		if v != "" {
			args = append(args, strings.Fields(v)...)
		}
	}

	username := "unknown user"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	wd := r.Form.Get("__work-dir__")
	if _, err := os.Stat(wd); wd == "" || os.IsNotExist(err) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Working directory %s does not exist", wd)))
		return
	}

	code, output, err := helper.CallExternalWithOutput([]string{}, wd, "cdt", args...)
	if r.Form.Has("__echo-cmd__") && r.Form.Get("__echo-cmd__") == "on" {
		w.Write([]byte(fmt.Sprintf("# %s in %s$ cdt %s\n", username, wd, strings.Join(args, " "))))
	}
	w.Write([]byte(output))
	if r.Form.Has("__show-exit-code__") && r.Form.Get("__show-exit-code__") == "on" {
		w.Write([]byte(fmt.Sprintf("\nexit code: %d", code)))
	}
}
