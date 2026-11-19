package uiapp

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// 底部栏只保留框架：不渲染根目录概览、统计数字或操作反馈。
func TestStatusBarHasNoContent(t *testing.T) {
	ta := newTestApp(t)
	ta.setStatus("图集 #3 已创建，请选择图片上传")
	tt := ui.NewTester(ta.View, 1200, 800)
	settled(tt)
	if err := tt.Click("图集列表"); err != nil {
		t.Fatalf("click 图集列表: %v", err)
	}
	tt.Frame()

	for _, text := range []string{"本地数据库", "根目录：", "SQLite", "最近操作", "状态：", "图集 #3 已创建，请选择图片上传"} {
		if tt.HasText(text) {
			t.Fatalf("bottom bar should be empty but shows %q; texts: %v", text, tt.Texts())
		}
	}
}
