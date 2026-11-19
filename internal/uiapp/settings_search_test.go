package uiapp

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSettingsSearchIndex(t *testing.T) {
	// 中文关键字与英文别名都应命中。
	for _, q := range []string{"语言", "language", "备份", "backup", "更新", "sqlite"} {
		results := settingsSearchResults(q)
		if len(results) == 0 {
			t.Errorf("query %q returned no results", q)
		}
	}
	// 无意义关键字应无结果。
	if got := settingsSearchResults("zzzznotfound"); len(got) != 0 {
		t.Errorf("nonsense query returned %d results", len(got))
	}
}

func TestSettingsPageMatchesFilter(t *testing.T) {
	if !settingsPageMatches("general", "语言") {
		t.Error("general page should match 语言")
	}
	if settingsPageMatches("about", "备份") {
		t.Error("about page should not match 备份")
	}
	// 页面标题本身命中时应展示整页。
	if !settingsPageMatches("backup", "备份") {
		t.Error("backup page should match its own title")
	}
}

func TestSettingsSearchResultsViewRenders(t *testing.T) {
	ta := newTestApp(t)
	ta.settingsQuery = "备份"
	tt := ui.NewTester(ta.settingsView, 780, 640)
	tt.Frame()

	texts := strings.Join(tt.Texts(), " | ")
	for _, want := range []string{"搜索结果", "立即备份", "定时备份", "备份周期"} {
		if !strings.Contains(texts, want) {
			t.Errorf("search results view missing %q; texts: %s", want, texts)
		}
	}
}

func TestSettingsNoResultsViewRenders(t *testing.T) {
	ta := newTestApp(t)
	ta.settingsQuery = "zzzznotfound"
	tt := ui.NewTester(ta.settingsView, 780, 640)
	tt.Frame()

	texts := strings.Join(tt.Texts(), " | ")
	if !strings.Contains(texts, "未找到相关设置") {
		t.Errorf("no-results view missing; texts: %s", texts)
	}
}
