package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"autocarsystem/internal/config"
	"autocarsystem/internal/media"
	"autocarsystem/internal/store"
)

func TestLocalBusinessFlow(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	paths := config.Paths{
		CacheTmp:   filepath.Join(dir, "data", "cache", "tmp"),
		MediaError: filepath.Join(dir, "data", "logs", "media-errors.jsonl"),
		Yuvision:   filepath.Join(dir, "media", "youvision"),
		Gallery:    filepath.Join(dir, "media", "gallery"),
	}
	for _, p := range []string{paths.CacheTmp, filepath.Dir(paths.MediaError), paths.Yuvision, paths.Gallery} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
	}

	st, err := store.OpenAndMigrate(ctx, filepath.Join(dir, "data", "car.db"), filepath.Join(dir, "data", "backup"))
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}
	defer st.Close()

	svc := New(st, paths)
	if err := svc.Bootstrap(ctx); err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	mustExec(t, ctx, st, "INSERT INTO brand(id, name) VALUES (1, '测试品牌')")
	mustExec(t, ctx, st, "INSERT INTO firm(id, brand_id, name) VALUES (1, 1, '测试厂商')")
	mustExec(t, ctx, st, "INSERT INTO series(id, brand_id, firm_id, name) VALUES (1, 1, 1, '测试车系')")

	yuvisionID, err := svc.CreateYuvision(ctx, 1, 1, 1, 1, "测试版型-01", "描述")
	if err != nil {
		t.Fatalf("create yuvision failed: %v", err)
	}
	if yuvisionID <= 0 {
		t.Fatalf("invalid yuvision id: %d", yuvisionID)
	}

	galleryID, err := svc.CreateGallery(ctx, yuvisionID, "测试图集-01")
	if err != nil {
		t.Fatalf("create gallery failed: %v", err)
	}

	if err := svc.DeleteYuvision(ctx, yuvisionID); err == nil {
		t.Fatal("expected yuvision delete restriction when gallery exists")
	}

	src := filepath.Join(dir, "input.jpg")
	if err := os.WriteFile(src, []byte("fake image content"), 0o644); err != nil {
		t.Fatalf("write source image failed: %v", err)
	}

	yvRel, err := svc.AddYuvisionImage(ctx, yuvisionID, src)
	if err != nil {
		t.Fatalf("add yuvision image failed: %v", err)
	}
	if !strings.HasPrefix(yvRel, "detail/") {
		t.Fatalf("unexpected yuvision rel path: %s", yvRel)
	}
	yvBase := filepath.Join(paths.Yuvision, "1", "1")
	yvFile := filepath.Join(yvBase, filepath.FromSlash(yvRel))
	if _, err := os.Stat(yvFile); err != nil {
		t.Fatalf("yuvision image file missing: %v", err)
	}
	yvThumb := filepath.Join(yvBase, filepath.FromSlash(media.ThumbnailRel(yvRel)))
	if _, err := os.Stat(yvThumb); err != nil {
		t.Fatalf("yuvision thumbnail missing: %v", err)
	}

	galRel, err := svc.AddGalleryImage(ctx, galleryID, src)
	if err != nil {
		t.Fatalf("add gallery image failed: %v", err)
	}
	if !strings.HasPrefix(galRel, "focus/") {
		t.Fatalf("unexpected gallery rel path: %s", galRel)
	}
	galBase := filepath.Join(paths.Gallery, "1", fmt.Sprintf("%d", galleryID))
	galThumb := filepath.Join(galBase, filepath.FromSlash(media.ThumbnailRel(galRel)))
	if _, err := os.Stat(galThumb); err != nil {
		t.Fatalf("gallery thumbnail missing: %v", err)
	}

	gallery, err := svc.getGallery(ctx, galleryID)
	if err != nil {
		t.Fatalf("get gallery failed: %v", err)
	}
	if gallery.ImageCount != 1 {
		t.Fatalf("unexpected image_count: %d", gallery.ImageCount)
	}
	if gallery.CoverImage != galRel {
		t.Fatalf("unexpected cover: %s", gallery.CoverImage)
	}

	if err := svc.RemoveGalleryImage(ctx, galleryID, galRel); err != nil {
		t.Fatalf("remove gallery image failed: %v", err)
	}
	gallery, err = svc.getGallery(ctx, galleryID)
	if err != nil {
		t.Fatalf("get gallery after remove failed: %v", err)
	}
	if gallery.ImageCount != 0 || gallery.CoverImage != "" {
		t.Fatalf("gallery image state not cleared: count=%d cover=%s", gallery.ImageCount, gallery.CoverImage)
	}

	if err := svc.DeleteGallery(ctx, galleryID); err != nil {
		t.Fatalf("delete gallery failed: %v", err)
	}
	if err := svc.DeleteYuvision(ctx, yuvisionID); err != nil {
		t.Fatalf("delete yuvision failed: %v", err)
	}
}

func mustExec(t *testing.T, ctx context.Context, st *store.Store, sql string, args ...any) {
	t.Helper()
	if _, err := st.Execute(ctx, sql, args...); err != nil {
		t.Fatalf("exec failed %s: %v", sql, err)
	}
}
