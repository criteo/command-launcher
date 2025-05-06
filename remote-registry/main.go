package main

import (
	"log"
	"net/http"

	remote "github.com/criteo/command-launcher/internal/remote"
	handlers "github.com/criteo/command-launcher/remote-registry/handlers"
	model "github.com/criteo/command-launcher/remote-registry/model"
	. "github.com/criteo/command-launcher/remote-registry/store"
)

func main() {
	// create a new in-memory store
	store := NewInMemoryStore()
	initStore(store)

	// create a controller instance by specifying the store
	controller := handlers.NewController(store)

	mux := http.NewServeMux()
	// Home Page
	mux.HandleFunc("/", controller.HomePageHandler)
	// GET the remote registry index
	mux.HandleFunc("/registry/{registry}/index.json", controller.IndexHandler)

	mux.HandleFunc("/registry", controller.RegistryHandler)
	mux.HandleFunc("/registry/{registry}", controller.RegistryHandler)
	mux.HandleFunc("/registry/{registry}/package", controller.PackageHandler)
	mux.HandleFunc("/registry/{registry}/package/{package}", controller.PackageHandler)
	mux.HandleFunc("/registry/{registry}/package/{package}/version", controller.PackageVersionHandler)
	mux.HandleFunc("/registry/{registry}/package/{package}/version/{version}", controller.PackageVersionHandler)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func initStore(store Store) {
	// put some example data in the store
	store.NewRegistry("test-registry", model.RegistryMetadata{
		Name:         "test-registry",
		Description:  "Test Registry",
		Admin:        []string{"a", "b"},
		CustomValues: map[string]string{"key": "value"},
	})
	// test-registry
	store.NewPackage("test-registry", "test-package", model.PackageMetadata{
		Name:         "test-package",
		Description:  "Test Package",
		Admin:        []string{"a", "b"},
		CustomValues: map[string]string{"key": "value"},
	})
	store.NewPackageVersion("test-registry", "test-package", "1.0.0", remote.PackageInfo{
		Name:           "test-package",
		Version:        "1.0.0",
		Url:            "http://example.com/test-package-1.0.0.tar.gz",
		Checksum:       "abc123",
		StartPartition: 0,
		EndPartition:   9,
	})
	store.NewPackageVersion("test-registry", "test-package", "1.1.0", remote.PackageInfo{
		Name:           "test-package",
		Version:        "1.1.0",
		Url:            "http://example.com/test-package-1.1.0.tar.gz",
		Checksum:       "abc456",
		StartPartition: 0,
		EndPartition:   9,
	})

	// test-registry-2
	store.NewRegistry("test-registry-2", model.RegistryMetadata{
		Name:         "test-registry-2",
		Description:  "Test Registry 2",
		Admin:        []string{"a", "b"},
		CustomValues: map[string]string{"key": "value"},
	})
	store.NewPackage("test-registry-2", "test-package-2", model.PackageMetadata{
		Name:         "test-package-2",
		Description:  "Test Package 2",
		Admin:        []string{"a", "b"},
		CustomValues: map[string]string{"key": "value"},
	})
	store.NewPackageVersion("test-registry-2", "test-package-2", "1.0.0", remote.PackageInfo{
		Name:           "test-package-2",
		Version:        "1.0.0",
		Url:            "http://example.com/test-package-2-1.0.0.tar.gz",
		Checksum:       "abc123",
		StartPartition: 0,
		EndPartition:   9,
	})

	regs, _ := store.AllRegistries()
	log.Printf("Store initialized with example data, number of registries: %d\n", len(regs))
}
