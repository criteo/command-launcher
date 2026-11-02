package helper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadProgressUsesStderr(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "package contents")
	}))
	defer server.Close()

	dir := t.TempDir()
	stdout, err := os.Create(filepath.Join(dir, "stdout"))
	require.NoError(t, err)
	defer stdout.Close()
	stderr, err := os.Create(filepath.Join(dir, "stderr"))
	require.NoError(t, err)
	defer stderr.Close()
	originalStdout, originalStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdout, stderr
	defer func() { os.Stdout, os.Stderr = originalStdout, originalStderr }()

	dest := filepath.Join(dir, "download")
	require.NoError(t, DownloadFileFromUrl(server.URL, dest, true))
	contents, err := os.ReadFile(dest)
	require.NoError(t, err)
	require.Equal(t, "package contents", string(contents))
	output, err := os.ReadFile(stdout.Name())
	require.NoError(t, err)
	require.Empty(t, output)
	diagnostics, err := os.ReadFile(stderr.Name())
	require.NoError(t, err)
	require.Contains(t, string(diagnostics), "Initializing download...")
	// stderr is a file here, so no cursor moves are written.
	require.NotContains(t, string(diagnostics), "\033[")
}
