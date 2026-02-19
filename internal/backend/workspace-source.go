package backend

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/criteo/command-launcher/internal/repository"
)

const WorkspaceSourcePrefix = "workspace:"

// WorkspacePackagesFileName returns the dot file name for the given app name.
// For example, app "cdt" -> ".cdt-packages", app "cola" -> ".cola-packages".
func WorkspacePackagesFileName(appName string) string {
	return fmt.Sprintf(".%s-packages", appName)
}

// DiscoverWorkspaceSources walks up from startDir to the filesystem root,
// looking for .<appName>-packages files. Returns workspace PackageSources ordered
// deepest-first (closest to startDir has highest priority).
func DiscoverWorkspaceSources(startDir string, appName string) []*PackageSource {
	sources := []*PackageSource{}
	dir := startDir
	checked := ""
	fileName := WorkspacePackagesFileName(appName)

	for dir != checked {
		candidate := filepath.Join(dir, fileName)
		if _, err := os.Stat(candidate); err == nil {
			pkgPaths, err := ParseWorkspaceFile(candidate)
			if err != nil {
				log.Warnf("workspace: failed to parse %s: %v", candidate, err)
			} else if len(pkgPaths) > 0 {
				src, err := NewWorkspaceSource(dir, pkgPaths)
				if err != nil {
					log.Warnf("workspace: failed to create source from %s: %v", candidate, err)
				} else {
					sources = append(sources, src)
				}
			}
		}
		checked = dir
		dir = filepath.Dir(dir)
	}

	return sources
}

// ParseWorkspaceFile reads a .cdt-packages file and returns absolute paths
// to valid package directories. Lines starting with # are comments.
// Absolute paths and paths containing ".." are rejected for security
// (packages must be under the workspace directory).
func ParseWorkspaceFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	baseDir := filepath.Dir(filePath)
	var paths []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// reject absolute paths and paths containing ".." for security
		if filepath.IsAbs(line) {
			log.Warnf("workspace: rejecting path %q in %s: absolute paths are not allowed", line, filePath)
			continue
		}
		if containsParentTraversal(line) {
			log.Warnf("workspace: rejecting path %q in %s: parent directory traversal (..) is not allowed", line, filePath)
			continue
		}

		absPath := filepath.Join(baseDir, line)

		// validate that the path exists and contains a manifest.mf
		manifestPath := filepath.Join(absPath, "manifest.mf")
		if _, err := os.Stat(manifestPath); err != nil {
			log.Warnf("workspace: skipping %q in %s: manifest.mf not found at %s", line, filePath, manifestPath)
			continue
		}

		paths = append(paths, absPath)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return paths, nil
}

// containsParentTraversal checks if a path contains ".." components.
func containsParentTraversal(path string) bool {
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

// NewWorkspaceSource creates a PackageSource for a workspace directory
// containing a .cdt-packages file.
func NewWorkspaceSource(workspaceDir string, packagePaths []string) (*PackageSource, error) {
	name := WorkspaceSourcePrefix + workspaceDir

	repoIndex, err := repository.NewWorkspaceRepoIndex(name, packagePaths)
	if err != nil {
		return nil, err
	}

	return &PackageSource{
		Name:            name,
		RepoDir:         workspaceDir,
		RemoteBaseURL:   "",
		SyncPolicy:      SYNC_POLICY_NEVER,
		IsManaged:       false,
		CustomRepoIndex: repoIndex,
	}, nil
}
