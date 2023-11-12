package server

import (
	"html/template"
	"net/http"
	"strings"
)

func (server *Server) CommandHandler(w http.ResponseWriter, r *http.Request) {
	fullName := strings.TrimPrefix(r.URL.Path, "/command/")
	cmd, err := (*(server.backend)).FindCommandByFullName(fullName)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	tmpl_text, _ := templates.ReadFile("templates/command.html")
	tmpl, _ := template.New("command").Parse(string(tmpl_text))
	command := Command{
		FullName: cmd.FullName(),
		Group:    cmd.Group(),
		Name:     cmd.Name(),
		Package:  cmd.PackageName(),
		Registry: cmd.RepositoryID(),
		Short:    cmd.ShortDescription(),
		Long:     cmd.LongDescription(),
		Examples: cmd.Examples(),
		Flags:    cmd.Flags(),
	}

	tmpl.Execute(w, command)
}
