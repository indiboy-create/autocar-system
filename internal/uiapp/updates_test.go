package uiapp

import (
	"testing"
	"time"

	"autocarsystem/internal/i18n"

	"github.com/egoist/mygo/ui"
)

func renderUpdates(t *testing.T, a *App) *ui.Tester {
	t.Helper()
	i18n.Set("zh-CN")
	a.settingsPage = "updates"
	a.updates = true
	tt := ui.NewTester(a.settingsView, 900, 700)
	tt.Frame()
	return tt
}

// fakeUpdater 把插件的偏好与检查函数换成内存实现，直到返回函数恢复它们。
func fakeUpdater(checks, downloads *bool, last *time.Time, checked *int) func() {
	check, getChecks, setChecks := checkForUpdates, updateChecks, setUpdateChecks
	getDownloads, setDownloads, getLast := updateDownloads, setUpdateDownloads, lastUpdateCheck
	checkForUpdates = func() { *checked++ }
	updateChecks = func() bool { return *checks }
	setUpdateChecks = func(on bool) { *checks = on }
	updateDownloads = func() bool { return *downloads }
	setUpdateDownloads = func(on bool) { *downloads = on }
	lastUpdateCheck = func() time.Time { return *last }
	return func() {
		checkForUpdates, updateChecks, setUpdateChecks = check, getChecks, setChecks
		updateDownloads, setUpdateDownloads, lastUpdateCheck = getDownloads, setDownloads, getLast
	}
}

// 与 mygo-music 一致：更新页控件不因构建能力禁用，只有自动下载依赖自动检查。
func TestUpdateControlsNotDisabledByBuild(t *testing.T) {
	checks, downloads, checked := true, false, 0
	var last time.Time
	defer fakeUpdater(&checks, &downloads, &last, &checked)()

	ta := newTestApp(t)
	tt := renderUpdates(t, ta.App)

	for label, role := range map[string]ui.Role{
		"自动检查更新":  ui.RoleSwitch,
		"自动下载并安装": ui.RoleSwitch,
		"检查更新":    ui.RoleButton,
	} {
		n, ok := tt.AccessByRole(role, label)
		if !ok {
			t.Fatalf("control %q not found", label)
		}
		if n.Disabled {
			t.Errorf("control %q should be enabled", label)
		}
	}
}

// 自动检查关闭时，自动下载开关随之禁用。
func TestTurnOffUpdateChecksDisablesDownloads(t *testing.T) {
	checks, downloads, checked := false, true, 0
	var last time.Time
	defer fakeUpdater(&checks, &downloads, &last, &checked)()

	ta := newTestApp(t)
	tt := renderUpdates(t, ta.App)

	downloadSwitch, ok := tt.AccessByRole(ui.RoleSwitch, "自动下载并安装")
	if !ok {
		t.Fatal("downloads switch not found")
	}
	if !downloadSwitch.Disabled {
		t.Fatal("downloads switch should be disabled while checks are off")
	}
}

// 上次检查时间以相对时间显示在自动检查的描述里，与 mygo-music 一致。
func TestUpdateCheckDescriptionUsesRelativeTime(t *testing.T) {
	checks, downloads, checked := true, false, 0
	last := time.Now().Add(-3 * 24 * time.Hour)
	defer fakeUpdater(&checks, &downloads, &last, &checked)()

	ta := newTestApp(t)
	tt := renderUpdates(t, ta.App)
	if !tt.HasText("3 天前") {
		t.Fatalf("relative check time not shown: %v", tt.Texts())
	}
}

// 设置页的检查按钮走插件接缝。
func TestCheckForUpdatesFromSettings(t *testing.T) {
	checks, downloads, checked := true, false, 0
	var last time.Time
	defer fakeUpdater(&checks, &downloads, &last, &checked)()

	ta := newTestApp(t)
	tt := renderUpdates(t, ta.App)
	if err := tt.Click("检查更新"); err != nil {
		t.Fatal(err)
	}
	if checked != 1 {
		t.Fatalf("checked %d times, want 1", checked)
	}
}

// 托盘检查更新同样走插件，未注册插件时不触发检查。
func TestTrayCheckUpdates(t *testing.T) {
	checks, downloads, checked := true, false, 0
	var last time.Time
	defer fakeUpdater(&checks, &downloads, &last, &checked)()

	ta := newTestApp(t)
	ta.updates = false
	ta.TrayCheckUpdates()
	if checked != 0 {
		t.Fatal("should not check when updater plugin is not used")
	}

	ta.updates = true
	ta.TrayCheckUpdates()
	if checked != 1 {
		t.Fatalf("tray checked %d times, want 1", checked)
	}
}

// 旧偏好只迁移一次；之后不再用 prefs 覆盖插件状态。
func TestMigrateUpdatePrefsRunsOnce(t *testing.T) {
	checks, downloads, checked := false, false, 0
	var last time.Time
	defer fakeUpdater(&checks, &downloads, &last, &checked)()

	ta := newTestApp(t)
	ta.prefs.AutoCheckUpdates = true
	ta.prefs.AutoInstallUpdates = true
	ta.prefs.UpdatePrefsMigrated = false

	ta.migrateUpdatePrefs()
	if !checks || !downloads {
		t.Fatalf("prefs not migrated: checks=%v downloads=%v", checks, downloads)
	}
	if !ta.prefs.UpdatePrefsMigrated {
		t.Fatal("migration flag not set")
	}

	setUpdateChecks(false)
	ta.migrateUpdatePrefs()
	if checks {
		t.Fatal("second migration overwrote plugin state")
	}
}
