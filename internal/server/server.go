package server

import (
	"embed"
	"fmt"
	"net/http"

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

type Command struct {
	FullName       string
	Group          string
	Name           string
	Package        string
	Registry       string
	Short          string
	Long           string
	Examples       []command.ExampleEntry
	Flags          []command.Flag
	SubCmds        []*Command
	DefaultWorkDir string
	HasAlias       bool
	Alias          string
}

type CommandIndex struct {
	Commands []*Command
}

func (server *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	index, _ := fs.ReadFile("static/index.html")
	w.Write(index)
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
	http.HandleFunc("/execute/", server.ExecuteHandler)
	http.HandleFunc("/health", server.HealthHandler)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
