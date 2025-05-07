package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	remote "github.com/criteo/command-launcher/internal/remote"
	handlers "github.com/criteo/command-launcher/remote-registry/handlers"
	model "github.com/criteo/command-launcher/remote-registry/model"
	. "github.com/criteo/command-launcher/remote-registry/store"
)

type CommandLineArgs struct {
	StoreType string
	StorePath string
	ShowHelp  bool
}

func setupCommandLineArgs() (*CommandLineArgs, error) {
	pflag.String("store", "memory", "Type of store to use: 'memory' or 'filesystem'")
	pflag.String("store-path", "", "Path for file store (required when using filesystem store)")
	pflag.Bool("help", false, "Display this message")
	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	viper.SetDefault("store", "memory")

	viper.SetEnvPrefix("REGISTRY")
	viper.AutomaticEnv()

	return &CommandLineArgs{
		StoreType: viper.GetString("store"),
		StorePath: viper.GetString("store-path"),
		ShowHelp:  viper.GetBool("help"),
	}, nil
}

func createStore(config *CommandLineArgs) (Store, error) {
	var store Store
	var err error

	switch config.StoreType {
	case "memory":
		log.Println("Using in-memory store")
		store = NewInMemoryStore()
	case "filesystem":
		storePath := config.StorePath
		if storePath == "" {
			log.Printf("No path specified, using default working directory")
			wd, err := os.Getwd()
			if err == nil {
				storePath = filepath.Join(wd, "registry-store")
			} else {
				return nil, fmt.Errorf("failed to get working directory: %w", err)
			}
			log.Printf("Using path: %s", storePath)
		}

		log.Printf("Using filesystem store at: %s", storePath)
		store, err = NewFileStore(storePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem store: %w", err)
		}
		log.Printf("Filesystem store created at %s", storePath)
	default:
		return nil, fmt.Errorf("unknown store type: %s", config.StoreType)
	}

	return store, nil
}

func main() {
	config, err := setupCommandLineArgs()
	if err != nil {
		log.Fatalf("Failed to set up configuration: %v", err)
	}

	if config.ShowHelp {
		pflag.PrintDefaults()
		os.Exit(0)
	}

	// Create the appropriate store
	store, err := createStore(config)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

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
