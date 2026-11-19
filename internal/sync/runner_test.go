package sync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"autocarsystem/internal/config"
	"autocarsystem/internal/service"
	"autocarsystem/internal/store"
)

func TestRunWithConfigParsesRealFieldShapes(t *testing.T) {
	brandJSON := `{
  "returncode": 0,
  "result": {
    "brand": [
      {
        "firstletter": "A",
        "brandlist": [
          {"id": 33, "name": "奥迪", "logo": "logo.png", "state": 1, "havenewenergy": 1}
        ]
      }
    ]
  }
}`

	seriesJSON := `{
  "returncode": 0,
  "result": {
    "brandid": 33,
    "fcttagdto": [
      {"id": 9, "name": "一汽奥迪", "link": "firm-link"}
    ],
    "groupList": [
      {
        "id": 1,
        "name": "轿车",
        "serieslist": [
          {
            "seriesid": 3170,
            "seriesName": "奥迪A3",
            "seriesState": 20,
            "seriesPriceMin": 165900,
            "seriesPriceMax": 209900,
            "fctid": 9,
            "fctname": "一汽奥迪",
            "seriesplace": "合资",
            "seriespy": "ADA3",
            "rank": 5,
            "pnglogo": "series.png",
            "newenergy": 0,
            "levelId": 3,
            "levelName": "紧凑型车",
            "isForeignCar": 0,
            "uvRank": 5,
            "hotRank": 134,
            "fuelnameList": ["汽油"],
            "istop": 0,
            "cartypename": null,
            "seriesTag1": null
          }
        ]
      }
    ]
  }
}`

	specJSON := `{
  "returncode": 0,
  "message": "",
  "result": [
    {
      "volumeGroupName": "1.5T 160马力 涡轮增压",
      "specList": [
        {
          "specName": "2026款 Sportback 35TFSI",
          "seriesName": "奥迪A3",
          "minOriginalPrice": 165900,
          "maxOriginalPrice": 165900,
          "imageUrl": "spec.jpg",
          "newsPrice": 165900,
          "newsId": 0,
          "dealerId": 0,
          "saleType": 0,
          "specWhiteImgUrl": "",
          "specId": 73985
        }
      ]
    }
  ]
}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/brand":
			_, _ = w.Write([]byte(brandJSON))
		case "/series":
			_, _ = w.Write([]byte(seriesJSON))
		case "/spec":
			_, _ = w.Write([]byte(specJSON))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	dir := t.TempDir()
	paths := config.Paths{
		StatePath: filepath.Join(dir, "sync", "state.json"),
		Conflicts: filepath.Join(dir, "sync", "conflicts.jsonl"),
		RawDir:    filepath.Join(dir, "sync", "raw"),
		DBPath:    filepath.Join(dir, "data", "car.db"),
		BackupDir: filepath.Join(dir, "data", "backup"),
	}
	for _, p := range []string{filepath.Dir(paths.StatePath), paths.RawDir, filepath.Dir(paths.DBPath)} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
	}

	st, err := store.OpenAndMigrate(ctx, paths.DBPath, paths.BackupDir)
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}
	defer st.Close()

	svc := service.New(st, paths)
	cfg := DefaultConfig()
	cfg.BrandURL = server.URL + "/brand"
	cfg.SeriesURLFormat = server.URL + "/series?brandid=%d"
	cfg.SpecURLFormat = server.URL + "/spec?seriesid=%d"
	cfg.Retry = 1
	cfg.Delay = 0
	cfg.SyncSpecs = true

	progressCount := 0
	if err := RunWithConfig(ctx, svc, paths, cfg, func(Progress) { progressCount++ }); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if progressCount == 0 {
		t.Fatal("expected progress callbacks")
	}

	counts, err := svc.Counts(ctx)
	if err != nil {
		t.Fatalf("counts failed: %v", err)
	}
	if counts["brand"] != 1 || counts["firm"] != 1 || counts["series"] != 1 || counts["spec"] != 1 {
		t.Fatalf("unexpected counts: %+v", counts)
	}

	specs, err := svc.ListSpecs(ctx, 10)
	if err != nil {
		t.Fatalf("list specs failed: %v", err)
	}
	if len(specs) != 1 || specs[0].ModelYear != "2026" {
		t.Fatalf("unexpected spec model year: %+v", specs)
	}

	state, err := LoadState(paths.StatePath)
	if err != nil {
		t.Fatalf("load state failed: %v", err)
	}
	if state.Stage != "done" || state.Mode != "full_scan" {
		t.Fatalf("unexpected state: %+v", state)
	}
}
