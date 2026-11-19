package store

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureSeededWritesSeedWhenMissing(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "data", "car.db")
	seed := []byte("prefilled-db")

	if err := EnsureSeeded(dbPath, seed); err != nil {
		t.Fatalf("ensure seeded failed: %v", err)
	}
	got, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("read seeded db failed: %v", err)
	}
	if !bytes.Equal(got, seed) {
		t.Fatalf("seeded db = %q, want %q", got, seed)
	}
}

func TestEnsureSeededSkipsExisting(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "car.db")
	existing := []byte("user-data")
	if err := os.WriteFile(dbPath, existing, 0o644); err != nil {
		t.Fatalf("write existing db failed: %v", err)
	}

	if err := EnsureSeeded(dbPath, []byte("different-seed")); err != nil {
		t.Fatalf("ensure seeded failed: %v", err)
	}
	got, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("read db failed: %v", err)
	}
	if !bytes.Equal(got, existing) {
		t.Fatalf("existing db was overwritten: got %q, want %q", got, existing)
	}
}

func TestOpenUsesExistingDatabaseWithoutCreatingSchema(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "car.db")

	st, err := OpenAndMigrate(ctx, dbPath, filepath.Join(dir, "backup"))
	if err != nil {
		t.Fatalf("prepare seeded db failed: %v", err)
	}
	if _, err := st.Execute(ctx, "INSERT INTO brand(id, name) VALUES (1, '预置品牌')"); err != nil {
		t.Fatalf("insert seed row failed: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	st, err = Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer st.Close()

	rows, err := st.Query(ctx, "SELECT name FROM brand WHERE id = 1")
	if err != nil {
		t.Fatalf("query seeded row failed: %v", err)
	}
	if len(rows.Values) != 1 {
		t.Fatalf("seeded data lost: got %d rows", len(rows.Values))
	}
}

func TestOpenRejectsMissingDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "missing.db")
	st, err := Open(context.Background(), dbPath)
	if err == nil {
		st.Close()
		t.Fatal("opening a missing database should fail instead of creating one")
	}
	if _, statErr := os.Stat(dbPath); !os.IsNotExist(statErr) {
		t.Fatal("Open must not create a database file")
	}
}

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
