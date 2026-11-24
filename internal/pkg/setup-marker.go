package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SETUP_MARKER_FILE is the name of the file written into a package directory
// once its __setup__ hook has completed successfully.
const SETUP_MARKER_FILE = ".setup"

// SetupMarker records that a package's setup phase has run.
type SetupMarker struct {
	CompletedAt    time.Time `json:"completedAt"`
	PackageVersion string    `json:"packageVersion"`
}

// IsSetupDone reports whether the setup marker exists in pkgDir.
// An empty pkgDir never counts as done (and is never looked up relative to the cwd).
func IsSetupDone(pkgDir string) bool {
	if pkgDir == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(pkgDir, SETUP_MARKER_FILE))
	return err == nil
}

// MarkSetupDone writes the setup marker file into pkgDir.
// It refuses an empty pkgDir so the marker can never land in the process cwd.
func MarkSetupDone(pkgDir string, packageVersion string) error {
	if pkgDir == "" {
		return fmt.Errorf("cannot write setup marker: empty package directory")
	}
	marker := SetupMarker{
		CompletedAt:    time.Now(),
		PackageVersion: packageVersion,
	}

	data, err := json.MarshalIndent(marker, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(pkgDir, SETUP_MARKER_FILE), data, 0644)
}
