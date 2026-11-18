package uiapp

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

// 所有文本输入框都有无障碍名称：Placeholder 不会产生名称，必须显式 Label。
func TestTextFieldsHaveAccessibleNames(t *testing.T) {
	// 目录页的编辑表单只在 catalogEditID != 0 时渲染，四种实体逐一覆盖。
	pages := []struct{ page, entity string }{
		{"gallery", ""}, {"youvision", ""}, {"seriesQuery", ""}, {"seriesFilter", ""},
		{"catalog", "brand"}, {"catalog", "firm"}, {"catalog", "series"}, {"catalog", "spec"},
	}
	for _, tc := range pages {
		name := tc.page
		if tc.entity != "" {
			name += "/" + tc.entity
		}
		t.Run(name, func(t *testing.T) {
			ta := newTestApp(t)
			ta.page = tc.page
			ta.catalogEntity = tc.entity
			ta.catalogEditID = 1 // 展开目录编辑表单
			tt := ui.NewTester(ta.View, 1200, 900)
			tt.Frame()

			fields := 0
			for _, info := range tt.AccessAll() {
				if info.Role == ui.RoleTextField {
					fields++
					if strings.TrimSpace(info.Name) == "" {
						t.Errorf("text field on %s has no accessible name (placeholder=%q)", name, info.Placeholder)
					}
				}
			}
			if fields == 0 {
				t.Errorf("no text fields found on %s", name)
			}
		})
	}
}
