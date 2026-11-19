package uiapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSavePrefsWithToastSetsSavedMessage(t *testing.T) {
	ta := newTestApp(t)
	ta.prefs.Language = "zh-CN"

	ta.savePrefsWithToast("界面语言")

	if ta.toast.Kind != "success" {
		t.Fatalf("toast kind = %q, want success", ta.toast.Kind)
	}
	if ta.toast.Text != "界面语言设置已保存" {
		t.Fatalf("toast text = %q, want 界面语言设置已保存", ta.toast.Text)
	}
	if _, err := os.Stat(ta.paths.PrefsPath); err != nil {
		t.Fatalf("prefs not written: %v", err)
	}
}

func TestSavePrefsWithToastReportsFailure(t *testing.T) {
	ta := newTestApp(t)
	blocker := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	ta.paths.PrefsPath = filepath.Join(blocker, "prefs.json")

	ta.savePrefsWithToast("界面语言")

	if ta.toast.Kind != "error" {
		t.Fatalf("toast kind = %q, want error", ta.toast.Kind)
	}
	if !strings.HasPrefix(ta.toast.Text, "设置保存失败：") {
		t.Fatalf("toast text = %q, want failure message", ta.toast.Text)
	}
}

func TestSettingsViewConsumesToast(t *testing.T) {
	ta := newTestApp(t)
	ta.settingsPage = "general"
	ta.toast = appToast{Text: "界面语言设置已保存", Kind: "success"}

	tt := ui.NewTester(ta.settingsView, 780, 640)
	tt.Frame()

	if ta.toast.Text != "" {
		t.Fatalf("toast not consumed by settings view: %+v", ta.toast)
	}
}

func TestToastColorsAdaptToTheme(t *testing.T) {
	successLight, _, _ := toastColors("success", false)
	successDark, _, _ := toastColors("success", true)
	errorLight, _, _ := toastColors("error", false)

	if successLight == successDark {
		t.Fatal("success background should differ between light and dark themes")
	}
	if successLight == errorLight {
		t.Fatal("success and error backgrounds should differ")
	}
}
