package config

import (
	"os"
	"path/filepath"
	"testing"
)

func useExecutable(t *testing.T, exe string) {
	t.Helper()
	orig := executable
	executable = func() (string, error) { return exe, nil }
	t.Cleanup(func() { executable = orig })
}

func TestLoadUsesExecutableDirectory(t *testing.T) {
	root := t.TempDir()
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	useExecutable(t, filepath.Join(root, "autoCarSystem.exe"))

	t.Setenv("AUTOCAR_ROOT", filepath.Join(root, "ignored"))

	paths, err := Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if paths.Root != root {
		t.Fatalf("unexpected root: %s, want %s", paths.Root, root)
	}
	if paths.Data != filepath.Join(root, "data") {
		t.Fatalf("unexpected data path: %s", paths.Data)
	}
	if paths.Media != filepath.Join(root, "media") {
		t.Fatalf("unexpected media path: %s", paths.Media)
	}
	if paths.DBPath != filepath.Join(root, "data", "car.db") {
		t.Fatalf("unexpected db path: %s", paths.DBPath)
	}

	if err := paths.EnsureDirs(); err != nil {
		t.Fatalf("ensure dirs failed: %v", err)
	}
	if _, err := os.Stat(paths.RawDir); err != nil {
		t.Fatalf("raw dir missing: %v", err)
	}
	if _, err := os.Stat(paths.Media); err != nil {
		t.Fatalf("media dir missing: %v", err)
	}
}

func TestLoadRejectsUnavailableExecutable(t *testing.T) {
	orig := executable
	executable = func() (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { executable = orig })

	if _, err := Load(); err == nil {
		t.Fatal("expected error when executable path is unavailable")
	}
}
