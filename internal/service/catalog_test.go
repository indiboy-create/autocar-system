package service

import (
	"context"
	"path/filepath"
	"testing"

	"autocarsystem/internal/config"
	"autocarsystem/internal/domain"
	"autocarsystem/internal/store"
)

func TestCatalogEditUpdatesOnlyNonIdentityFields(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	st, err := store.OpenAndMigrate(ctx, filepath.Join(dir, "car.db"), filepath.Join(dir, "backup"))
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}
	defer st.Close()

	mustExec(t, ctx, st, "INSERT INTO brand(id, name) VALUES (1, '原品牌')")
	mustExec(t, ctx, st, "INSERT INTO firm(id, brand_id, name) VALUES (1, 1, '原厂商')")
	mustExec(t, ctx, st, "INSERT INTO series(id, brand_id, firm_id, name) VALUES (1, 1, 1, '原车系')")
	mustExec(t, ctx, st, "INSERT INTO spec(id, brand_id, firm_id, series_id, name) VALUES (1, 1, 1, 1, '原年款')")

	svc := New(st, config.Paths{})

	if err := svc.UpdateBrand(ctx, 1, domain.BrandEdit{
		Name: "新品牌", FirstLetter: "X", State: 1, HaveNewEnergy: 1, APISort: 9,
	}); err != nil {
		t.Fatalf("update brand failed: %v", err)
	}
	if err := svc.UpdateFirm(ctx, 1, domain.FirmEdit{Name: "新厂商", PY: "xcs"}); err != nil {
		t.Fatalf("update firm failed: %v", err)
	}
	if err := svc.UpdateSeries(ctx, 1, domain.SeriesEdit{
		Name: "新车系", GroupName: "组", State: 1, PriceMin: 1.5, PriceMax: 9.5,
	}); err != nil {
		t.Fatalf("update series failed: %v", err)
	}
	if err := svc.UpdateSpec(ctx, 1, domain.SpecEdit{
		Name: "新年款", ModelYear: "2026", PriceMin: 1, PriceMax: 2, NewsPrice: 1.5,
	}); err != nil {
		t.Fatalf("update spec failed: %v", err)
	}

	assertStringValue(t, ctx, st, "SELECT name FROM brand WHERE id = 1", "新品牌")
	assertStringValue(t, ctx, st, "SELECT name FROM firm WHERE id = 1", "新厂商")
	assertStringValue(t, ctx, st, "SELECT name FROM series WHERE id = 1", "新车系")
	assertStringValue(t, ctx, st, "SELECT model_year FROM spec WHERE id = 1", "2026")
	assertIntValue(t, ctx, st, "SELECT brand_id FROM firm WHERE id = 1", 1)
	assertIntValue(t, ctx, st, "SELECT series_id FROM spec WHERE id = 1", 1)
	assertNonEmpty(t, ctx, st, "SELECT user_updated_at FROM brand WHERE id = 1")
}

func assertStringValue(t *testing.T, ctx context.Context, st *store.Store, sql string, want string) {
	t.Helper()
	rows, err := st.Query(ctx, sql)
	if err != nil {
		t.Fatalf("query failed %s: %v", sql, err)
	}
	if len(rows.Values) == 0 || len(rows.Values[0]) == 0 {
		t.Fatalf("no value for %s", sql)
	}
	got := valueString(rows.Values[0][0])
	if got != want {
		t.Fatalf("unexpected value for %s: got %q, want %q", sql, got, want)
	}
}

func assertIntValue(t *testing.T, ctx context.Context, st *store.Store, sql string, want int64) {
	t.Helper()
	rows, err := st.Query(ctx, sql)
	if err != nil {
		t.Fatalf("query failed %s: %v", sql, err)
	}
	if len(rows.Values) == 0 || len(rows.Values[0]) == 0 {
		t.Fatalf("no value for %s", sql)
	}
	got := valueInt64(rows.Values[0][0])
	if got != want {
		t.Fatalf("unexpected value for %s: got %d, want %d", sql, got, want)
	}
}

func assertNonEmpty(t *testing.T, ctx context.Context, st *store.Store, sql string) {
	t.Helper()
	rows, err := st.Query(ctx, sql)
	if err != nil {
		t.Fatalf("query failed %s: %v", sql, err)
	}
	if len(rows.Values) == 0 || len(rows.Values[0]) == 0 || valueString(rows.Values[0][0]) == "" {
		t.Fatalf("expected non-empty value for %s", sql)
	}
}
