package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"autocarsystem/internal/config"
	"autocarsystem/internal/domain"
	"autocarsystem/internal/store"
)

func TestSearchGalleriesAndScanOrphans(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	paths := config.Paths{
		Gallery:    filepath.Join(dir, "media", "gallery"),
		MediaError: filepath.Join(dir, "data", "logs", "media-errors.jsonl"),
	}
	if err := os.MkdirAll(paths.Gallery, 0o755); err != nil {
		t.Fatalf("create gallery dir failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(paths.MediaError), 0o755); err != nil {
		t.Fatalf("create log dir failed: %v", err)
	}

	st, err := store.OpenAndMigrate(ctx, filepath.Join(dir, "data", "car.db"), filepath.Join(dir, "data", "backup"))
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}
	defer st.Close()

	mustExec(t, ctx, st, "INSERT INTO brand(id, name) VALUES (1, '品牌')")
	mustExec(t, ctx, st, "INSERT INTO firm(id, brand_id, name) VALUES (1, 1, '厂商')")
	mustExec(t, ctx, st, "INSERT INTO series(id, brand_id, firm_id, name) VALUES (1, 1, 1, '车系')")
	mustExec(t, ctx, st, "INSERT INTO category(id, name, sort, enabled) VALUES (1, '分类', 0, 1)")
	mustExec(t, ctx, st, "INSERT INTO youvision(id, brand_id, firm_id, series_id, category_id, name, created_at, updated_at) VALUES (1, 1, 1, 1, 1, '版型', '', '')")
	mustExec(t, ctx, st, "INSERT INTO gallery(id, youvision_id, brand_id, firm_id, series_id, category_id, title, created_at, updated_at) VALUES (1, 1, 1, 1, 1, 1, 'Alpha', '2026-01-01', '2026-01-01')")
	mustExec(t, ctx, st, "INSERT INTO gallery(id, youvision_id, brand_id, firm_id, series_id, category_id, title, created_at, updated_at) VALUES (2, 1, 1, 1, 1, 1, 'Beta', '2026-01-02', '2026-01-02')")

	svc := New(st, paths)

	found, err := svc.SearchGalleries(ctx, domain.GalleryFilter{Keyword: "Alpha"}, 20)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(found) != 1 || found[0].Title != "Alpha" {
		t.Fatalf("unexpected search result: %+v", found)
	}

	sorted, err := svc.SearchGalleries(ctx, domain.GalleryFilter{Sort: "title_asc"}, 20)
	if err != nil {
		t.Fatalf("sort failed: %v", err)
	}
	if len(sorted) < 2 || sorted[0].Title != "Alpha" {
		t.Fatalf("unexpected sort result: %+v", sorted)
	}

	knownDir := filepath.Join(paths.Gallery, "1", "1")
	orphanDir := filepath.Join(paths.Gallery, "1", "999")
	if err := os.MkdirAll(knownDir, 0o755); err != nil {
		t.Fatalf("create known dir failed: %v", err)
	}
	if err := os.MkdirAll(orphanDir, 0o755); err != nil {
		t.Fatalf("create orphan dir failed: %v", err)
	}

	orphans, err := svc.ScanOrphanGalleryDirs(ctx)
	if err != nil {
		t.Fatalf("scan orphans failed: %v", err)
	}
	if len(orphans) != 1 || orphans[0] != "1/999" {
		t.Fatalf("unexpected orphans: %+v", orphans)
	}

	removed, err := svc.CleanupOrphanGalleryDirs(ctx)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if removed != 1 {
		t.Fatalf("unexpected removed count: %d", removed)
	}
	if _, err := os.Stat(orphanDir); !os.IsNotExist(err) {
		t.Fatalf("orphan dir still exists: %v", err)
	}
	if _, err := os.Stat(knownDir); err != nil {
		t.Fatalf("known dir was removed: %v", err)
	}
}
