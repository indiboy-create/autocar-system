package uiapp

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"autocarsystem/internal/i18n"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func renderUpdates(t *testing.T, a *App) *ui.Tester {
	t.Helper()
	i18n.Set("zh-CN")
	a.settingsPage = "updates"
	tt := ui.NewTester(a.settingsView, 900, 700)
	tt.Frame()
	return tt
}

// waitMessage 等待 checkUpdateNow 的后台结果落到 settingsMessage。
func waitMessage(t *testing.T, a *App, want string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if a.settingsMessage == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("message = %q, want %q", a.settingsMessage, want)
}

// 未配置更新能力时：明确说明且控件禁用，不假装可用。
func TestUpdatesDisabledIsExplicit(t *testing.T) {
	oldEnabled := updatesEnabled
	updatesEnabled = func() bool { return false }
	t.Cleanup(func() { updatesEnabled = oldEnabled })

	ta := newTestApp(t)
	tt := renderUpdates(t, ta.App)

	if !tt.HasText("当前构建未启用自动更新") {
		t.Fatalf("disabled notice not shown: %v", tt.Texts())
	}
	for label, role := range map[string]ui.Role{
		"自动检查更新":  ui.RoleSwitch,
		"自动下载并安装": ui.RoleSwitch,
		"检查更新":    ui.RoleButton,
	} {
		n, ok := tt.AccessByRole(role, label)
		if !ok {
			t.Fatalf("control %q not found", label)
		}
		if !n.Disabled {
			t.Errorf("control %q should be disabled when updates are off", label)
		}
	}
}

// 已配置更新能力时：检查按钮走真实接缝，并把结果显示在消息区。
func TestUpdatesCheckUsesSeam(t *testing.T) {
	oldEnabled, oldCheck := updatesEnabled, checkForUpdate
	updatesEnabled = func() bool { return true }
	checkForUpdate = func(context.Context) (*mygo.Update, error) {
		return &mygo.Update{Version: "9.9.9"}, nil
	}
	t.Cleanup(func() { updatesEnabled, checkForUpdate = oldEnabled, oldCheck })

	ta := newTestApp(t)
	ta.settingsMessage = ""
	ta.checkUpdateNow()
	waitMessage(t, ta.App, "发现新版本 9.9.9")
}

// 已是最新版本时给出明确提示。
func TestUpdatesCheckUpToDate(t *testing.T) {
	oldEnabled, oldCheck := updatesEnabled, checkForUpdate
	updatesEnabled = func() bool { return true }
	checkForUpdate = func(context.Context) (*mygo.Update, error) { return nil, nil }
	t.Cleanup(func() { updatesEnabled, checkForUpdate = oldEnabled, oldCheck })

	ta := newTestApp(t)
	ta.settingsMessage = ""
	ta.checkUpdateNow()
	waitMessage(t, ta.App, "已是最新版本")
}

// 检查失败时给出可读错误，而不是静默。
func TestUpdatesCheckError(t *testing.T) {
	oldEnabled, oldCheck := updatesEnabled, checkForUpdate
	updatesEnabled = func() bool { return true }
	checkForUpdate = func(context.Context) (*mygo.Update, error) { return nil, errors.New("boom") }
	t.Cleanup(func() { updatesEnabled, checkForUpdate = oldEnabled, oldCheck })

	ta := newTestApp(t)
	ta.settingsMessage = ""
	ta.checkUpdateNow()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && !strings.HasPrefix(ta.settingsMessage, "检查更新失败") {
		time.Sleep(5 * time.Millisecond)
	}
	if !strings.HasPrefix(ta.settingsMessage, "检查更新失败") {
		t.Fatalf("message = %q, want error prefix", ta.settingsMessage)
	}
}

// 未配置时按钮点击只给说明，不触发检查接缝。
func TestUpdatesDisabledDoesNotCheck(t *testing.T) {
	oldEnabled, oldCheck := updatesEnabled, checkForUpdate
	called := false
	updatesEnabled = func() bool { return false }
	checkForUpdate = func(context.Context) (*mygo.Update, error) { called = true; return nil, nil }
	t.Cleanup(func() { updatesEnabled, checkForUpdate = oldEnabled, oldCheck })

	ta := newTestApp(t)
	ta.checkUpdateNow()
	if called {
		t.Fatal("check should not run when updates are disabled")
	}
	if ta.settingsMessage != "当前构建未启用自动更新" {
		t.Fatalf("message = %q", ta.settingsMessage)
	}
}

// 未开启“自动检查更新”时，启动流程不请求更新服务。
func TestAutoCheckSkipsWhenOff(t *testing.T) {
	oldEnabled, oldCheck := updatesEnabled, checkForUpdate
	called := make(chan struct{}, 1)
	updatesEnabled = func() bool { return true }
	checkForUpdate = func(context.Context) (*mygo.Update, error) {
		called <- struct{}{}
		return nil, nil
	}
	t.Cleanup(func() { updatesEnabled, checkForUpdate = oldEnabled, oldCheck })

	ta := newTestApp(t)
	ta.prefs.AutoCheckUpdates = false
	ta.AutoCheckUpdates()
	select {
	case <-called:
		t.Fatal("check should not run when automatic checks are off")
	case <-time.After(100 * time.Millisecond):
	}
}

// 每天只自动检查一次：刚检查过就跳过。
func TestAutoCheckSkipsWhenRecentlyChecked(t *testing.T) {
	oldEnabled, oldCheck := updatesEnabled, checkForUpdate
	called := make(chan struct{}, 1)
	updatesEnabled = func() bool { return true }
	checkForUpdate = func(context.Context) (*mygo.Update, error) {
		called <- struct{}{}
		return nil, nil
	}
	t.Cleanup(func() { updatesEnabled, checkForUpdate = oldEnabled, oldCheck })

	ta := newTestApp(t)
	ta.prefs.AutoCheckUpdates = true
	ta.prefs.LastUpdateCheck = time.Now().UTC().Format(time.RFC3339)
	ta.AutoCheckUpdates()
	select {
	case <-called:
		t.Fatal("check should be throttled to once a day")
	case <-time.After(100 * time.Millisecond):
	}
}

// 超过一天没有检查时启动检查，并把新的发现写进更新设置页。
func TestAutoCheckRunsDailyAndReports(t *testing.T) {
	oldEnabled, oldCheck := updatesEnabled, checkForUpdate
	updatesEnabled = func() bool { return true }
	checkForUpdate = func(context.Context) (*mygo.Update, error) {
		return &mygo.Update{Version: "9.9.9"}, nil
	}
	t.Cleanup(func() { updatesEnabled, checkForUpdate = oldEnabled, oldCheck })

	ta := newTestApp(t)
	ta.prefs.AutoCheckUpdates = true
	ta.prefs.LastUpdateCheck = time.Now().Add(-25 * time.Hour).UTC().Format(time.RFC3339)
	old := ta.prefs.LastUpdateCheck
	ta.AutoCheckUpdates()
	waitMessage(t, ta.App, "发现新版本 9.9.9")
	if ta.prefs.LastUpdateCheck == old {
		t.Fatal("last update check time was not recorded")
	}
}

// 开启“自动下载并安装”时，后台检查到新版本后直接安装。
func TestAutoCheckInstallsWhenEnabled(t *testing.T) {
	oldEnabled, oldCheck, oldInstall := updatesEnabled, checkForUpdate, installUpdate
	updatesEnabled = func() bool { return true }
	checkForUpdate = func(context.Context) (*mygo.Update, error) {
		return &mygo.Update{Version: "9.9.9"}, nil
	}
	installed := make(chan string, 1)
	installUpdate = func(_ context.Context, u *mygo.Update) error {
		installed <- u.Version
		return nil
	}
	t.Cleanup(func() { updatesEnabled, checkForUpdate, installUpdate = oldEnabled, oldCheck, oldInstall })

	ta := newTestApp(t)
	ta.prefs.AutoCheckUpdates = true
	ta.prefs.AutoInstallUpdates = true
	ta.AutoCheckUpdates()
	waitMessage(t, ta.App, "新版本 9.9.9 已下载，下次启动时安装")
	select {
	case v := <-installed:
		if v != "9.9.9" {
			t.Fatalf("installed = %q", v)
		}
	default:
		t.Fatal("update was not installed")
	}
}
