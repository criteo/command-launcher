package server

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"

	. "github.com/criteo/command-launcher/internal/backend"
	"github.com/criteo/command-launcher/internal/command"
)

type Server struct {
	backend *Backend
}

//go:embed static
var fs embed.FS

//go:embed templates
var templates embed.FS

func (server *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	index, _ := fs.ReadFile("static/index.html")
	w.Write(index)
}

type Command struct {
	FullName string
	Group    string
	Name     string
	Package  string
	Registry string
	Short    string
	Long     string
	Examples []command.ExampleEntry
	Flags    []command.Flag
	SubCmds  []*Command
}

type CommandIndex struct {
	Commands []*Command
}

func (server *Server) CommandIndexHandler(w http.ResponseWriter, r *http.Request) {
	repos := (*(server.backend)).AllRepositories()
	cmdMap := map[string]*Command{}

	for _, repo := range repos {
		pkgs := repo.InstalledPackages()
		for _, pkg := range pkgs {
			cmds := pkg.Commands()
			for _, cmd := range cmds {
				if cmd.Name() == "__setup__" {
					continue
				}
				if cmd.Group() == "" { // this is top level command
					// add group command
					if groupCmd, ok := cmdMap[cmd.FullName()]; !ok {
						groupCmd = &Command{
							FullName: cmd.FullName(),
							Group:    cmd.Group(),
							Name:     cmd.Name(),
							Package:  pkg.Name(),
							Registry: repo.Name(),
							SubCmds:  []*Command{},
						}
						cmdMap[cmd.FullName()] = groupCmd
					}
				} else { // this is a sub command
					// get the group command full name
					groupCmdFullName := fmt.Sprintf("%s@@%s@%s", cmd.Group(), pkg.Name(), repo.Name())
					if groupCmd, ok := cmdMap[groupCmdFullName]; !ok {
						groupCmd = &Command{
							FullName: groupCmdFullName,
							Group:    "",
							Name:     cmd.Group(),
							Package:  pkg.Name(),
							Registry: repo.Name(),
							SubCmds: []*Command{
								&Command{
									FullName: cmd.FullName(),
									Group:    cmd.Group(),
									Name:     cmd.Name(),
									Package:  pkg.Name(),
									Registry: repo.Name(),
									SubCmds:  []*Command{},
								},
							},
						}
						cmdMap[groupCmdFullName] = groupCmd
					} else {
						groupCmd.SubCmds = append(groupCmd.SubCmds, &Command{
							FullName: cmd.FullName(),
							Group:    cmd.Group(),
							Name:     cmd.Name(),
							Package:  pkg.Name(),
							Registry: repo.Name(),
							SubCmds:  []*Command{},
						})
					}

				}
			}
		}
	}

	// load the template
	tmpl_text, _ := templates.ReadFile("templates/cmd_index.html")
	tmpl, _ := template.New("command_index").Parse(string(tmpl_text))

	groups := make([]*Command, 0, len(cmdMap))
	for _, cmd := range cmdMap {
		groups = append(groups, cmd)
	}

	sort.Slice(groups, func(i, j int) bool {
		// return fmt.Sprintf("%s@%s", groups[i].Package, groups[i].Registry) < fmt.Sprintf("%s@%s", groups[j].Package, groups[j].Registry)
		return groups[i].FullName < groups[j].FullName
	})

	var tpl bytes.Buffer
	tmpl.Execute(&tpl, CommandIndex{Commands: groups})

	html_wrapper, _ := templates.ReadFile("templates/html_wrapper.html")
	content := strings.ReplaceAll(string(html_wrapper), "#INPUT#", tpl.String())
	w.Write([]byte(content))
}

func (server *Server) CommandHandler(w http.ResponseWriter, r *http.Request) {
	fullName := strings.TrimPrefix(r.URL.Path, "/command/")
	fmt.Println(r.URL.Path, fullName)
	cmd, _ := (*(server.backend)).FindCommandByFullName(fullName)

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

func (server *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// Implement the serve funcation in default backend
func Serve(backend *Backend, port int) error {
	server := Server{backend: backend}

	http.HandleFunc("/", server.CommandIndexHandler)
	http.HandleFunc("/test", server.CommandIndexHandler)
	http.HandleFunc("/command/", server.CommandHandler)
	http.HandleFunc("/health", server.HealthHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
