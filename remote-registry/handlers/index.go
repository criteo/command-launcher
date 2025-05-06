package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/criteo/command-launcher/internal/remote"
	. "github.com/criteo/command-launcher/remote-registry/store"
)

type Controller struct {
	store Store
}

func NewController(store Store) *Controller {
	return &Controller{
		store: store,
	}
}

func (c Controller) HomePageHandler(w http.ResponseWriter, r *http.Request) {
	regs, _ := c.store.AllRegistries()
	names := make([]string, 0)
	for _, reg := range regs {
		names = append(names, reg.Name)
	}
	output := strings.Join(names, ", ")
	w.Write(fmt.Appendf(nil, "Available registries: %s", output))
}

// GET /registry/{registry}/index.json
func (c Controller) IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the registry name from the URL
	registryName := r.PathValue("registry")
	if registryName == "" {
		http.Error(w,
			"Registry name is required",
			http.StatusBadRequest)
		return
	}
	pkgs, err := c.store.AllPackagesFromRegistry(registryName)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to get packages from registry %s: %s", registryName, err),
			http.StatusInternalServerError)
		return
	}
	allVersions := make([]remote.PackageInfo, 0)
	for _, pkg := range pkgs {
		allVersions = append(allVersions, pkg.Versions...)
	}
	b, err := json.Marshal(allVersions)
	w.Write(b)
}

func (c Controller) RegistryHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Registry Handler"))
}

func (c Controller) PackageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Package Handler"))
}

func (c Controller) PackageVersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Package Version Handler"))
}
