package uiapp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"autocarsystem/internal/config"
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/service"
	"autocarsystem/internal/store"

	"github.com/egoist/mygo/ui"
)

// testApp 把 App 与其底层 store 放在一起：store 仅供测试插入夹具数据，
// 生产代码只通过 service 访问它。
type testApp struct {
	*App
	st *store.Store
}

// newTestApp 构造一个带真实 SQLite store 的 App，不含窗口与事件循环。
// runOnMain 被替换为同步执行，使 Refresh 的快照阶段在测试 goroutine 内完成，
// 而不会等待无人服务的 mygo 主线程事件循环。
func newTestApp(t *testing.T) *testApp {
	t.Helper()
	i18n.Set("zh-CN") // 断言使用中文原文，避免依赖系统首选语言

	ctx := context.Background()
	dir := t.TempDir()
	paths := config.Paths{
		Root:       dir,
		Data:       filepath.Join(dir, "data"),
		Media:      filepath.Join(dir, "media"),
		DBPath:     filepath.Join(dir, "data", "car.db"),
		PrefsPath:  filepath.Join(dir, "data", "prefs.json"),
		BackupDir:  filepath.Join(dir, "data", "backup"),
		CacheTmp:   filepath.Join(dir, "data", "cache", "tmp"),
		Yuvision:   filepath.Join(dir, "media", "youvision"),
		Gallery:    filepath.Join(dir, "media", "gallery"),
		MediaError: filepath.Join(dir, "data", "logs", "media-errors.jsonl"),
	}
	for _, p := range []string{
		paths.Data, paths.BackupDir, paths.CacheTmp,
		filepath.Dir(paths.MediaError), paths.Yuvision, paths.Gallery,
	} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
	}

	st, err := store.OpenAndMigrate(ctx, paths.DBPath, paths.BackupDir)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	svc := service.New(st, paths)
	if err := svc.Bootstrap(ctx); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	a := &App{}
	a.svc = svc
	a.paths = paths
	a.page = "home"
	a.gallerySort = "updated_desc"
	a.catalogEntity = "brand"
	a.coverBitmaps = make(map[string]*ui.Bitmap)
	a.runOnMain = func(fn func()) { fn() }

	ta := &testApp{App: a, st: st}
	// 初始 Refresh：无窗口，走同步路径，填充分类/计数等状态。
	a.Refresh()
	return ta
}

// exec 直接执行一条 SQL，供测试插入夹具数据。
func (ta *testApp) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := ta.st.Execute(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

// settled 渲染一帧并推进动画，让视图处于静止状态。
func settled(tt *ui.Tester) {
	tt.Frame()
	tt.Advance(1e9) // 1s in ns
	tt.Frame()
}
