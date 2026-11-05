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

// isFileInUseError reports the errors returned while another process (a cdt
// reader or writer, an antivirus, the indexer) has the file open: by
// MoveFileEx when replacing it, and by CreateFile when opening it during a
// replace.
func isFileInUseError(err error) bool {
	return errors.Is(err, errorSharingViolation) || errors.Is(err, syscall.ERROR_ACCESS_DENIED)
}

// syncDir is a no-op: directories cannot be fsynced on Windows, and NTFS
// journals the rename metadata.
func syncDir(dir string) error {
	return nil
}
