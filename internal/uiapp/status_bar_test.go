package uiapp

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// 底部信息栏参考 mygo-music 的布局：状态/路径在左，数据统计在中，本地状态在右。
func TestStatusBarShowsStats(t *testing.T) {
	ta := newTestApp(t)
	ta.status = ""
	ta.counts = map[string]int64{
		"brand": 3, "firm": 5, "series": 7, "youvision": 11, "gallery": 13,
	}
	tt := ui.NewTester(ta.View, 1200, 800)
	tt.Frame()

	for _, text := range []string{"品牌", "厂商", "车系", "版型", "图集", "3", "5", "7", "11", "13", "SQLite", "本地数据库"} {
		if !tt.HasText(text) {
			t.Fatalf("status bar missing %q; texts: %v", text, tt.Texts())
		}
	}
}

func TestStatusBarShowsRecentActivityLabel(t *testing.T) {
	ta := newTestApp(t)
	ta.setStatus("图集 #3 已创建，请选择图片上传")
	tt := ui.NewTester(ta.View, 1200, 800)
	tt.Frame()

	if !tt.HasText("最近操作") {
		t.Fatalf("recent activity label not visible; texts: %v", tt.Texts())
	}
}
