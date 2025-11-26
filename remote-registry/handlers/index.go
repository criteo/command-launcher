package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/criteo/command-launcher/internal/remote"
	"github.com/criteo/command-launcher/remote-registry/model"
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
	// for now directly return the registry json
	b, err := json.Marshal(regs)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to get registries: %s", err),
			http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
	w.Write(b)
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

func (c Controller) NewRegistryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling NewRegistryHandler")
	if r.Method != http.MethodPost {
		http.Error(w,
			"Method not allowed",
			http.StatusMethodNotAllowed)
		return
	}

	// get registry metadata from request body
	var registryInfo model.RegistryMetadata
	if err := json.NewDecoder(r.Body).Decode(&registryInfo); err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to decode request body: %s", err),
			http.StatusBadRequest)
		return
	}

	err := c.store.NewRegistry(registryInfo.Name, registryInfo)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to create registry %s: %s", registryInfo.Name, err),
			http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(fmt.Appendf(nil, "Registry %s created", registryInfo.Name))
}

func (c Controller) UpdateOrDeleteRegistryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling UpdateOrDeleteRegistryHandler")
	if r.Method != http.MethodPut && r.Method != http.MethodDelete {
		http.Error(w,
			"Method not allowed",
			http.StatusMethodNotAllowed)
		return
	}

	registryName := r.PathValue("registry")
	log.Println("Registry name:", registryName)
	if registryName == "" {
		http.Error(w,
			"Registry name is required",
			http.StatusBadRequest)
		return
	}

	// Handle Delete
	if r.Method == http.MethodDelete {
		err := c.store.DeleteRegistry(registryName)
		if err != nil {
			http.Error(w,
				fmt.Sprintf("Failed to delete registry %s: %s", registryName, err),
				http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		w.Write(fmt.Appendf(nil, "Registry %s deleted", registryName))
		return
	}

	// Handle update
	var registryInfo model.RegistryMetadata
	if err := json.NewDecoder(r.Body).Decode(&registryInfo); err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to decode request body: %s", err),
			http.StatusBadRequest)
		return
	}
	if registryName != registryInfo.Name {
		http.Error(w,
			fmt.Sprintf("Registry name in URL %s does not match registry name in body %s", registryName, registryInfo.Name),
			http.StatusBadRequest)
		return
	}
	err := c.store.UpdateRegistry(registryName, registryInfo)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to update registry %s: %s", registryName, err),
			http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Appendf(nil, "Registry %s updated", registryName))
}

func (c Controller) NewPackageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling NewPackageHandler")
	if r.Method != http.MethodPost {
		http.Error(w,
			"Method not allowed",
			http.StatusMethodNotAllowed)
		return
	}
	// get registry name from URL
	registryName := r.PathValue("registry")
	if registryName == "" {
		http.Error(w,
			"Registry name is required",
			http.StatusBadRequest)
		return
	}
	// get package metadata from request body
	var packageInfo model.PackageMetadata
	if err := json.NewDecoder(r.Body).Decode(&packageInfo); err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to decode request body: %s", err),
			http.StatusBadRequest)
		return
	}
	if packageInfo.Name == "" {
		http.Error(w,
			"Package name is required",
			http.StatusBadRequest)
		return
	}
	err := c.store.NewPackage(registryName, packageInfo.Name, packageInfo)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to create package %s in registry %s: %s", packageInfo.Name, registryName, err),
			http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("New package created"))
}

func (c Controller) UpdateOrDeletePackageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling UpdateOrDeletePackageHandler")
	if r.Method != http.MethodPut && r.Method != http.MethodDelete {
		http.Error(w,
			"Method not allowed",
			http.StatusMethodNotAllowed)
		return
	}
	// get registry name from URL
	registryName := r.PathValue("registry")
	if registryName == "" {
		http.Error(w,
			"Registry name is required",
			http.StatusBadRequest)
		return
	}
	// get package name from URL
	packageName := r.PathValue("package")
	if packageName == "" {
		http.Error(w,
			"Package name is required",
			http.StatusBadRequest)
		return
	}
	// Handle Delete
	if r.Method == http.MethodDelete {
		err := c.store.DeletePackage(registryName, packageName)
		if err != nil {
			http.Error(w,
				fmt.Sprintf("Failed to delete package %s in registry %s: %s", packageName, registryName, err),
				http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("Package deleted"))
		return
	}

	// Handle update
	var packageInfo model.PackageMetadata
	if err := json.NewDecoder(r.Body).Decode(&packageInfo); err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to decode request body: %s", err),
			http.StatusBadRequest)
		return
	}
	if packageName != packageInfo.Name {
		http.Error(w,
			fmt.Sprintf("Package name in URL %s does not match package name in body %s", packageName, packageInfo.Name),
			http.StatusBadRequest)
		return
	}
	err := c.store.UpdatePackage(registryName, packageName, packageInfo)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to update package %s in registry %s: %s", packageName, registryName, err),
			http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Package updated"))
}

func (c Controller) NewPackageVersionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling NewPackageVersionHandler")
	if r.Method != http.MethodPost {
		http.Error(w,
			"Method not allowed",
			http.StatusMethodNotAllowed)
		return
	}

	// get registry name from URL
	registryName := r.PathValue("registry")
	if registryName == "" {
		http.Error(w,
			"Registry name is required",
			http.StatusBadRequest)
		return
	}
	// get package name from URL
	packageName := r.PathValue("package")
	if packageName == "" {
		http.Error(w,
			"Package name is required",
			http.StatusBadRequest)
		return
	}

	// get package info from request body
	var packageInfo remote.PackageInfo
	if err := json.NewDecoder(r.Body).Decode(&packageInfo); err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to decode request body: %s", err),
			http.StatusBadRequest)
		return
	}

	// TODO: make further checks on packageInfo, ex, validate URL, checksum, version format etc.
	if packageName != packageInfo.Name {
		http.Error(w,
			fmt.Sprintf("Package name in URL %s does not match package name in body %s", packageName, packageInfo.Name),
			http.StatusBadRequest)
		return
	}
	if packageInfo.Version == "" {
		http.Error(w,
			"Package version is required",
			http.StatusBadRequest)
		return
	}
	err := c.store.NewPackageVersion(registryName, packageName, packageInfo.Version, packageInfo)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to create package version %s for package %s in registry %s: %s", packageInfo.Version, packageName, registryName, err),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("New package version created"))
}

func (c Controller) DeletePackageVersionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling DeletePackageVersionHandler")
	if r.Method != http.MethodDelete {
		http.Error(w,
			"Method not allowed",
			http.StatusMethodNotAllowed)
		return
	}

	// get registry name from URL
	registryName := r.PathValue("registry")
	if registryName == "" {
		http.Error(w,
			"Registry name is required",
			http.StatusBadRequest)
		return
	}
	// get package name from URL
	packageName := r.PathValue("package")
	if packageName == "" {
		http.Error(w,
			"Package name is required",
			http.StatusBadRequest)
		return
	}
	// get package version from URL
	packageVersion := r.PathValue("version")
	if packageVersion == "" {
		http.Error(w,
			"Package version is required",
			http.StatusBadRequest)
		return
	}

	err := c.store.DeletePackageVersion(registryName, packageName, packageVersion)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to delete package version %s for package %s in registry %s: %s", packageVersion, packageName, registryName, err),
			http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("Delete Package Version Handler"))
}
