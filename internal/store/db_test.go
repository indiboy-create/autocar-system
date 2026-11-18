package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenAndMigrateCreatesSchema(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "car.db")
	backupDir := filepath.Join(dir, "backup")

	st, err := OpenAndMigrate(ctx, dbPath, backupDir)
	if err != nil {
		t.Fatalf("open and migrate failed: %v", err)
	}
	defer st.Close()

	version, err := readSchemaVersion(ctx, st.db)
	if err != nil {
		t.Fatalf("read version failed: %v", err)
	}
	if version != SchemaVersion {
		t.Fatalf("unexpected version: got %d, want %d", version, SchemaVersion)
	}

	rows, err := st.Query(ctx, "SELECT name FROM sqlite_master WHERE type = 'table' AND name IN ('brand','firm','series','spec','category','youvision','gallery') ORDER BY name")
	if err != nil {
		t.Fatalf("query tables failed: %v", err)
	}
	if len(rows.Values) != 7 {
		t.Fatalf("unexpected table count: got %d, want 7", len(rows.Values))
	}
}

func TestOpenAndMigrateIsIdempotent(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "car.db")
	backupDir := filepath.Join(dir, "backup")

	st, err := OpenAndMigrate(ctx, dbPath, backupDir)
	if err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	st, err = OpenAndMigrate(ctx, dbPath, backupDir)
	if err != nil {
		t.Fatalf("second open failed: %v", err)
	}
	defer st.Close()

	version, err := readSchemaVersion(ctx, st.db)
	if err != nil {
		t.Fatalf("read version failed: %v", err)
	}
	if version != SchemaVersion {
		t.Fatalf("unexpected version: got %d, want %d", version, SchemaVersion)
	}
}
