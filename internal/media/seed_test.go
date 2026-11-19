package media

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestEnsureSeededCreatesStructureAndDoesNotOverwrite(t *testing.T) {
	root := t.TempDir()
	seed := fstest.MapFS{
		"resources/media/autocar/brand/titlepic.png": &fstest.MapFile{Data: []byte("seed-brand")},
		"resources/media/gallery/titlepic.png":       &fstest.MapFile{Data: []byte("seed-gallery")},
		"resources/media/youvision/titlepic.png":     &fstest.MapFile{Data: []byte("seed-youvision")},
	}

	if err := EnsureSeeded(root, seed); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	for _, dir := range []string{"autocar/brand", "gallery", "youvision"} {
		if st, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir))); err != nil || !st.IsDir() {
			t.Fatalf("directory %s missing: %v", dir, err)
		}
	}
	brand := filepath.Join(root, "autocar", "brand", "titlepic.png")
	if data, err := os.ReadFile(brand); err != nil || string(data) != "seed-brand" {
		t.Fatalf("unexpected seeded brand image: %q, err %v", data, err)
	}

	if err := os.WriteFile(brand, []byte("user-image"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSeeded(root, seed); err != nil {
		t.Fatalf("second seed failed: %v", err)
	}
	if data, err := os.ReadFile(brand); err != nil || string(data) != "user-image" {
		t.Fatalf("existing file was overwritten: %q, err %v", data, err)
	}
}
