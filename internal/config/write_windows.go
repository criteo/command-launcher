//go:build windows

package config

import (
	"errors"
	"os"
	"syscall"
)

const errorSharingViolation syscall.Errno = 32

func chownLike(f *os.File, like os.FileInfo) error {
	return nil
}

// isTransientRenameError reports the errors MoveFileEx returns while another
// process (a cdt reader, an antivirus, the indexer) has the target open.
func isTransientRenameError(err error) bool {
	return errors.Is(err, errorSharingViolation) || errors.Is(err, syscall.ERROR_ACCESS_DENIED)
}

// syncDir is a no-op: directories cannot be fsynced on Windows, and NTFS
// journals the rename metadata.
func syncDir(dir string) error {
	return nil
}
