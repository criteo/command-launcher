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
	StoreType       string
	StorePath       string
	S3Bucket        string
	S3Region        string
	S3Endpoint      string
	S3AccessKey     string
	S3SecretKey     string
	S3UsePathStyle  bool
	ShowHelp        bool
}

func setupCommandLineArgs() (*CommandLineArgs, error) {
	// Define CLI flags
	pflag.String("store", "memory", "Type of store to use: 'memory', 'filesystem', or 's3'")
	pflag.String("store-path", "", "Path for file store (required when using filesystem store)")
	pflag.String("s3-bucket", "", "S3 bucket name (required when using s3 store)")
	pflag.String("s3-region", "us-east-1", "S3 region")
	pflag.String("s3-endpoint", "", "S3 endpoint URL (for S3-compatible services like MinIO)")
	pflag.String("s3-access-key", "", "S3 access key ID (optional, uses IAM role if not set)")
	pflag.String("s3-secret-key", "", "S3 secret access key (optional, uses IAM role if not set)")
	pflag.Bool("s3-use-path-style", false, "Use path-style S3 URLs (required for MinIO)")
	pflag.String("config", "", "Path to configuration file")
	pflag.Bool("help", false, "Display this message")
	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Set defaults
	viper.SetDefault("store", "memory")

	// Support environment variables
	viper.SetEnvPrefix("REGISTRY")
	viper.AutomaticEnv()

	// Config file support
	configureConfigFile()

	return &CommandLineArgs{
		StoreType:      viper.GetString("store"),
		StorePath:      viper.GetString("store-path"),
		S3Bucket:       viper.GetString("s3-bucket"),
		S3Region:       viper.GetString("s3-region"),
		S3Endpoint:     viper.GetString("s3-endpoint"),
		S3AccessKey:    viper.GetString("s3-access-key"),
		S3SecretKey:    viper.GetString("s3-secret-key"),
		S3UsePathStyle: viper.GetBool("s3-use-path-style"),
		ShowHelp:       viper.GetBool("help"),
	}, nil
}

// configureConfigFile sets up and loads the config file
func configureConfigFile() {
	// 1. Check if explicit config path is specified via flag or env var
	if configPath := viper.GetString("config"); configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// 2. Set config name and type for default search
		viper.SetConfigName("registry-config")
		viper.SetConfigType("yaml")

		// 3. Look for config file in these paths (in order)
		// Current directory
		viper.AddConfigPath(".")
		// User's home directory
		if homeDir, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(filepath.Join(homeDir, ".command-launcher"))
		}
		// System-wide config
		viper.AddConfigPath("/etc/command-launcher")
	}

	// 4. Read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error occurred
			log.Printf("Error reading config file: %v", err)
		} else {
			// No config file found, using only flags and environment variables
			log.Printf("No config file found, using only flags and environment variables")
		}
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
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
	case "s3":
		if config.S3Bucket == "" {
			return nil, fmt.Errorf("s3-bucket is required when using s3 store")
		}

		log.Printf("Using S3 store with bucket: %s, region: %s", config.S3Bucket, config.S3Region)
		if config.S3Endpoint != "" {
			log.Printf("Using custom S3 endpoint: %s", config.S3Endpoint)
		}

		s3Config := S3StoreConfig{
			Bucket:          config.S3Bucket,
			Region:          config.S3Region,
			Endpoint:        config.S3Endpoint,
			AccessKeyID:     config.S3AccessKey,
			SecretAccessKey: config.S3SecretKey,
			UsePathStyle:    config.S3UsePathStyle,
		}

		store, err = NewS3Store(s3Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 store: %w", err)
		}
		log.Printf("S3 store created successfully")
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

	// POST create a new registry
	mux.HandleFunc("/registry", controller.NewRegistryHandler)
	// PUT or DELETE update or delete a registry
	mux.HandleFunc("/registry/{registry}", controller.UpdateOrDeleteRegistryHandler)

	// POST create a new package
	mux.HandleFunc("/registry/{registry}/package", controller.NewPackageHandler)
	// PUT or DELETE update or delete a package
	mux.HandleFunc("/registry/{registry}/package/{package}", controller.UpdateOrDeletePackageHandler)

	// POST create a new package version
	mux.HandleFunc("/registry/{registry}/package/{package}/version", controller.NewPackageVersionHandler)
	// DELETE delete a package version
	mux.HandleFunc("/registry/{registry}/package/{package}/version/{version}", controller.DeletePackageVersionHandler)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// for dev purpose
// TODO: remove this
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
		Url:            "file:///tmp/test-package-1.0.0.pkg",
		StartPartition: 0,
		EndPartition:   9,
	})
	store.NewPackageVersion("test-registry", "test-package", "1.1.0", remote.PackageInfo{
		Name:           "test-package",
		Version:        "1.1.0",
		Url:            "file:///tmp/test-package-1.1.0.pkg",
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
		Url:            "file:///tmp/test-package-2-1.0.0.pkg",
		StartPartition: 0,
		EndPartition:   9,
	})

	regs, _ := store.AllRegistries()
	log.Printf("Store initialized with example data, number of registries: %d\n", len(regs))
}
