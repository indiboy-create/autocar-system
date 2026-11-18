package fileutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteFileAtomicOverwriteAndNoResidue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prefs.json")
	if err := WriteFileAtomic(path, []byte("first"), 0o644); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := WriteFileAtomic(path, []byte("second"), 0o644); err != nil {
		t.Fatalf("second write: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "second" {
		t.Fatalf("content = %q, want second", data)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "prefs.json" {
		t.Fatalf("atomic write left residue: %v", entries)
	}
}

func TestWriteFileAtomicPerm(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows 不实现 POSIX 权限位，Chmod 无意义")
	}
	path := filepath.Join(t.TempDir(), "secret.json")
	if err := WriteFileAtomic(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perm = %o, want 600", info.Mode().Perm())
	}
}
