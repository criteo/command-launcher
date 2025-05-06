package model

import (
	remote "github.com/criteo/command-launcher/internal/remote"
)

type RegistryMetadata struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Admin        []string          `json:"admin"`
	CustomValues map[string]string `json:"customValues"`
}

type Registry struct {
	RegistryMetadata
	Packages map[string]Package `json:"packages"`
}

type PackageMetadata struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Admin        []string          `json:"admin"`
	CustomValues map[string]string `json:"customValues"`
}

type Package struct {
	PackageMetadata
	Versions []remote.PackageInfo `json:"versions"`
}
