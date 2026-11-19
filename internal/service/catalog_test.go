package service

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"autocarsystem/internal/config"
	"autocarsystem/internal/store"
)

func TestListBrandsPage(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	st, err := store.OpenAndMigrate(ctx, filepath.Join(dir, "car.db"), filepath.Join(dir, "backup"))
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}
	defer st.Close()

	for i := 1; i <= 250; i++ {
		mustExec(t, ctx, st, "INSERT INTO brand(id, name) VALUES (?, ?)", i, fmt.Sprintf("品牌%d", i))
	}
	svc := New(st, config.Paths{})

	first, more, err := svc.ListBrandsPage(ctx, 0, 200)
	if err != nil {
		t.Fatalf("first page failed: %v", err)
	}
	if len(first) != 200 || !more {
		t.Fatalf("first page len=%d more=%v, want 200/true", len(first), more)
	}
	if first[0].ID != 1 || first[199].ID != 200 {
		t.Fatalf("first page IDs %d..%d, want 1..200", first[0].ID, first[199].ID)
	}

	second, more, err := svc.ListBrandsPage(ctx, 200, 200)
	if err != nil {
		t.Fatalf("second page failed: %v", err)
	}
	if len(second) != 50 || more {
		t.Fatalf("second page len=%d more=%v, want 50/false", len(second), more)
	}
	if second[0].ID != 201 || second[49].ID != 250 {
		t.Fatalf("second page IDs %d..%d, want 201..250", second[0].ID, second[49].ID)
	}
}

func TestSearchBrandsPage(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	st, err := store.OpenAndMigrate(ctx, filepath.Join(dir, "car.db"), filepath.Join(dir, "backup"))
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}
	defer st.Close()

	// 前 60 条普通名，后 100 条带相同关键词，验证能搜到首屏之外的品牌。
	for i := 1; i <= 160; i++ {
		name := fmt.Sprintf("品牌%d", i)
		if i > 60 {
			name = fmt.Sprintf("目标%d号", i)
		}
		mustExec(t, ctx, st, "INSERT INTO brand(id, name, firstletter) VALUES (?, ?, ?)", i, name, "T")
	}
	svc := New(st, config.Paths{})

	// 名称模糊搜索：命中第 61..160 共 100 条，分两页。
	first, more, err := svc.SearchBrandsPage(ctx, "目标", 0, 60)
	if err != nil {
		t.Fatalf("search page 0 failed: %v", err)
	}
	if len(first) != 60 || !more {
		t.Fatalf("search page 0 len=%d more=%v, want 60/true", len(first), more)
	}
	if first[0].ID != 61 {
		t.Fatalf("search page 0 first ID=%d, want 61", first[0].ID)
	}

	second, more, err := svc.SearchBrandsPage(ctx, "目标", 60, 60)
	if err != nil {
		t.Fatalf("search page 1 failed: %v", err)
	}
	if len(second) != 40 || more {
		t.Fatalf("search page 1 len=%d more=%v, want 40/false", len(second), more)
	}

	// 首字母搜索：100 条 "T"，取第一页即可。
	byLetter, _, err := svc.SearchBrandsPage(ctx, "T", 0, 60)
	if err != nil {
		t.Fatalf("search by letter failed: %v", err)
	}
	if len(byLetter) != 60 {
		t.Fatalf("search by letter len=%d, want 60", len(byLetter))
	}

	// 精确 ID 搜索：直接命中第 150 条。
	byID, more, err := svc.SearchBrandsPage(ctx, "150", 0, 60)
	if err != nil {
		t.Fatalf("search by ID failed: %v", err)
	}
	if len(byID) != 1 || more || byID[0].ID != 150 {
		t.Fatalf("search by ID = %v more=%v, want [150] false", byID, more)
	}
}

func TestSearchOtherEntitiesPage(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	st, err := store.OpenAndMigrate(ctx, filepath.Join(dir, "car.db"), filepath.Join(dir, "backup"))
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}
	defer st.Close()
	svc := New(st, config.Paths{})

	mustExec(t, ctx, st, "INSERT INTO brand(id, name) VALUES (1, '品牌1')")

	// 厂商：共 100 条，前 60 普通名，后 40 带关键词；验证搜到首屏之外并分页。
	for i := 1; i <= 100; i++ {
		name := fmt.Sprintf("厂商%d", i)
		if i > 60 {
			name = fmt.Sprintf("目标厂商%d", i)
		}
		mustExec(t, ctx, st, "INSERT INTO firm(id, brand_id, name, py) VALUES (?, 1, ?, ?)", i, name, "MB")
	}
	// 车型：共 100 条，后 40 带关键词。
	for i := 1; i <= 100; i++ {
		name := fmt.Sprintf("车型%d", i)
		if i > 60 {
			name = fmt.Sprintf("目标车型%d", i)
		}
		mustExec(t, ctx, st, "INSERT INTO series(id, brand_id, firm_id, name, py) VALUES (?, 1, 1, ?, ?)", i, name, "MX")
	}
	// 年款：共 100 条，后 40 带关键词，需挂在已有车型下。
	for i := 1; i <= 100; i++ {
		name := fmt.Sprintf("年款%d", i)
		if i > 60 {
			name = fmt.Sprintf("目标年款%d", i)
		}
		mustExec(t, ctx, st, "INSERT INTO spec(id, brand_id, firm_id, series_id, name, model_year) VALUES (?, 1, 1, 1, ?, ?)", i, name, "2026")
	}

	firms, more, err := svc.SearchFirmsPage(ctx, "目标厂商", 0, 60)
	if err != nil || len(firms) != 40 || more || firms[0].ID != 61 {
		t.Fatalf("SearchFirmsPage len=%d more=%v err=%v, want 40/false/61-first", len(firms), more, err)
	}
	series, more, err := svc.SearchSeriesPage(ctx, "目标车型", 0, 60)
	if err != nil || len(series) != 40 || more || series[0].ID != 61 {
		t.Fatalf("SearchSeriesPage len=%d more=%v err=%v, want 40/false/61-first", len(series), more, err)
	}
	specs, more, err := svc.SearchSpecsPage(ctx, "目标年款", 0, 60)
	if err != nil || len(specs) != 40 || more || specs[0].ID != 61 {
		t.Fatalf("SearchSpecsPage len=%d more=%v err=%v, want 40/false/61-first", len(specs), more, err)
	}

	// 精确 ID 搜索：三者都应只返回一条。
	firms, _, err = svc.SearchFirmsPage(ctx, "80", 0, 60)
	if err != nil || len(firms) != 1 || firms[0].ID != 80 {
		t.Fatalf("SearchFirmsPage by ID = %v err=%v, want [80]", firms, err)
	}
	series, _, err = svc.SearchSeriesPage(ctx, "80", 0, 60)
	if err != nil || len(series) != 1 || series[0].ID != 80 {
		t.Fatalf("SearchSeriesPage by ID = %v err=%v, want [80]", series, err)
	}
	specs, _, err = svc.SearchSpecsPage(ctx, "80", 0, 60)
	if err != nil || len(specs) != 1 || specs[0].ID != 80 {
		t.Fatalf("SearchSpecsPage by ID = %v err=%v, want [80]", specs, err)
	}
}

func TestListFirmsPage(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	st, err := store.OpenAndMigrate(ctx, filepath.Join(dir, "car.db"), filepath.Join(dir, "backup"))
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}
	defer st.Close()

	mustExec(t, ctx, st, "INSERT INTO brand(id, name) VALUES (1, '品牌1')")
	for i := 1; i <= 250; i++ {
		mustExec(t, ctx, st, "INSERT INTO firm(id, brand_id, name) VALUES (?, 1, ?)", i, fmt.Sprintf("厂商%d", i))
	}
	svc := New(st, config.Paths{})

	first, more, err := svc.ListFirmsPage(ctx, 0, 200)
	if err != nil {
		t.Fatalf("first page failed: %v", err)
	}
	if len(first) != 200 || !more {
		t.Fatalf("first page len=%d more=%v, want 200/true", len(first), more)
	}
	if first[0].ID != 1 || first[199].ID != 200 {
		t.Fatalf("first page IDs %d..%d, want 1..200", first[0].ID, first[199].ID)
	}

	second, more, err := svc.ListFirmsPage(ctx, 200, 200)
	if err != nil {
		t.Fatalf("second page failed: %v", err)
	}
	if len(second) != 50 || more {
		t.Fatalf("second page len=%d more=%v, want 50/false", len(second), more)
	}
	if second[0].ID != 201 || second[49].ID != 250 {
		t.Fatalf("second page IDs %d..%d, want 201..250", second[0].ID, second[49].ID)
	}
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
