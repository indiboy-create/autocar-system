package i18n

import "testing"

func TestLanguageSwitch(t *testing.T) {
	Set("zh-CN")
	if got := T("设置"); got != "设置" {
		t.Fatalf("Chinese: got %q", got)
	}

	Set("en-US")
	if got := T("设置"); got != "Settings" {
		t.Fatalf("English: got %q", got)
	}
	if got := T("车型查询"); got != "Model Search" {
		t.Fatalf("English: got %q", got)
	}
	if got := T("未知文案"); got != "未知文案" {
		t.Fatalf("untranslated text should stay as-is, got %q", got)
	}

	Set("en-US")
	if got := T("自动检查更新"); got != "Check for Updates Automatically" {
		t.Fatalf("English: got %q", got)
	}
}
