package server

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"sort"
	"strings"
)

func (server *Server) CommandIndexHandler(w http.ResponseWriter, r *http.Request) {
	repos := (*(server.backend)).AllRepositories()
	cmdMap := map[string]*Command{}

	for _, repo := range repos {
		pkgs := repo.InstalledPackages()
		for _, pkg := range pkgs {
			cmds := pkg.Commands()
			for _, cmd := range cmds {
				// exclude the system commands from the ui
				if slices.Contains([]string{"__setup__", "__login__"}, cmd.Name()) {
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
		return groups[i].FullName < groups[j].FullName
	})

	var tpl bytes.Buffer
	tmpl.Execute(&tpl, CommandIndex{Commands: groups})

	html_wrapper, _ := templates.ReadFile("templates/html_wrapper.html")
	content := strings.ReplaceAll(string(html_wrapper), "@TITLE@", "Command Index")
	content = strings.ReplaceAll(content, "@INPUT@", tpl.String())

	w.Write([]byte(content))
}
