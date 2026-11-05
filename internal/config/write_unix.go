//go:build unix

package config

import (
	"os"
	"syscall"
)

// chownLike gives f the owner of the file described by like. It only acts
// when running as root (e.g. under sudo): otherwise the replaced config would
// become root-owned and no longer writable by its user.
func chownLike(f *os.File, like os.FileInfo) error {
	if os.Geteuid() != 0 {
		return nil
	}
	st, ok := like.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}
	return f.Chown(int(st.Uid), int(st.Gid))
}
