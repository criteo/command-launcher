//go:build !unix

package config

import "os"

func chownLike(f *os.File, like os.FileInfo) error {
	return nil
}
