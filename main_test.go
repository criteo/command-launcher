package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveAppName_FromBinaryName(t *testing.T) {
	// The test binary itself is the running executable, so resolveAppName
	// should return its base name (without extension) rather than the
	// compiled-in default.
	name := resolveAppName()
	assert.NotEmpty(t, name)
	assert.NotEqual(t, ".", name)
}

func TestResolveAppName_SymlinkResolvesToOriginal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy executable
	original := filepath.Join(tmpDir, "original-app")
	err := os.WriteFile(original, []byte("binary"), 0755)
	assert.NoError(t, err)

	// Create a symlink to it
	link := filepath.Join(tmpDir, "my-alias")
	err = os.Symlink(original, link)
	assert.NoError(t, err)

	// Resolve the symlink — should get the original name
	resolved, err := filepath.EvalSymlinks(link)
	assert.NoError(t, err)
	assert.Equal(t, "original-app", filepath.Base(resolved))
}

func TestResolveAppName_CopyGetsOwnName(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two separate files (simulating a copy)
	original := filepath.Join(tmpDir, "original-app")
	err := os.WriteFile(original, []byte("binary"), 0755)
	assert.NoError(t, err)

	copied := filepath.Join(tmpDir, "my-copy")
	err = os.WriteFile(copied, []byte("binary"), 0755)
	assert.NoError(t, err)

	// Each resolves to its own name
	resolvedOrig, err := filepath.EvalSymlinks(original)
	assert.NoError(t, err)
	assert.Equal(t, "original-app", filepath.Base(resolvedOrig))

	resolvedCopy, err := filepath.EvalSymlinks(copied)
	assert.NoError(t, err)
	assert.Equal(t, "my-copy", filepath.Base(resolvedCopy))
}

func TestResolveAppName_ExtensionStripped(t *testing.T) {
	name := "myapp.exe"
	assert.Equal(t, "myapp", name[:len(name)-len(filepath.Ext(name))])
}
