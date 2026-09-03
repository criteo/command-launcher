package pkg

import (
	"encoding/json"
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
func IsSetupDone(pkgDir string) bool {
	_, err := os.Stat(filepath.Join(pkgDir, SETUP_MARKER_FILE))
	return err == nil
}

// MarkSetupDone writes the setup marker file into pkgDir.
func MarkSetupDone(pkgDir string, packageVersion string) error {
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
