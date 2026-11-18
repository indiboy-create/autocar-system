package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrefsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "prefs.json")
	want := Prefs{Language: "en", Appearance: "dark", DefaultPage: "gallery", AutoCheckUpdates: true}

	if err := SavePrefs(path, want); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := LoadPrefs(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got != want {
		t.Fatalf("round trip = %+v, want %+v", got, want)
	}
}

func TestSavePrefsOverwritesAtomicallyWithoutResidue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prefs.json")
	first := DefaultPrefs()
	first.Language = "zh-CN"
	if err := SavePrefs(path, first); err != nil {
		t.Fatal(err)
	}
	second := DefaultPrefs()
	second.Language = "en"
	if err := SavePrefs(path, second); err != nil {
		t.Fatal(err)
	}

	got, err := LoadPrefs(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Language != "en" {
		t.Fatalf("language = %q, want en", got.Language)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("atomic save left residue: %v", entries)
	}
}

func TestLoadPrefsMissingReturnsDefaults(t *testing.T) {
	got, err := LoadPrefs(filepath.Join(t.TempDir(), "none.json"))
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if got != DefaultPrefs() {
		t.Fatalf("got %+v, want defaults", got)
	}
}
