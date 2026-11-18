package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	// 首次写入。
	if err := writeFileAtomic(path, []byte("first"), 0o644); err != nil {
		t.Fatalf("first write: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read first: %v", err)
	}
	if string(got) != "first" {
		t.Fatalf("first content = %q, want %q", got, "first")
	}

	// 覆盖写入（Windows 上目标已存在，验证 rename 覆盖逻辑）。
	if err := writeFileAtomic(path, []byte("second"), 0o644); err != nil {
		t.Fatalf("second write: %v", err)
	}
	got, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("read second: %v", err)
	}
	if string(got) != "second" {
		t.Fatalf("second content = %q, want %q", got, "second")
	}

	// 不应残留临时文件或备份文件。
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "state.json" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("unexpected files left behind: %v", names)
	}
}

func TestSaveStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s := NewState()
	s.LastSyncAt = "2026-01-01 00:00:00"
	s.Stage = "brand"
	if err := SaveState(path, s); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := LoadState(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.LastSyncAt != s.LastSyncAt || loaded.Stage != s.Stage {
		t.Fatalf("round trip mismatch: %+v", loaded)
	}
}

func TestSaveStateRejectsNonDirTemp(t *testing.T) {
	// 目标路径的父目录是一个文件，MkdirAll 应报错而不是静默成功。
	dir := t.TempDir()
	parent := filepath.Join(dir, "notadir")
	if err := os.WriteFile(parent, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := SaveState(filepath.Join(parent, "state.json"), NewState())
	if err == nil || !strings.Contains(err.Error(), "创建同步状态目录失败") {
		t.Fatalf("expected mkdir error, got %v", err)
	}
}
