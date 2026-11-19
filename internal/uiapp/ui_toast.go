package uiapp

import (
	"time"

	myui "github.com/egoist/mygo/ui"
)

// toastColors 按提示类型与明暗主题返回卡片的背景、边框与文字颜色。
// 成功用绿色系、失败用红色系，与系统的浅色/深色外观一致。
func toastColors(kind string, dark bool) (bg, border, text myui.Color) {
	base := statusSuccessColor
	if kind == "error" {
		base = statusDangerColor
	}
	if dark {
		return base.Alpha(0.18), base.Alpha(0.5), base
	}
	return base.Alpha(0.12), base.Alpha(0.35), base
}

// toastViewport 用自定义外观替换 MyGo 默认 toast：右下角圆角卡片，
// 上滑淡入、下滑淡出。显示时长、悬停暂停与堆叠仍由框架管理。
func (a *App) toastViewport(c *myui.Context) {
	myui.ToastViewportBase(c, func(viewport myui.Element, toasts []myui.Toast) {
		viewport.Bottom(24).Right(24).AlignItems(myui.End).Gap(8)
		for _, t := range toasts {
			toast := myui.ToastBase(c, t)
			bg, border, text := toastColors(t.Type, c.Theme().Dark)
			box := toast.Root.Row().
				MinHeight(40).MaxWidth(360).Padding(11, 16).Radius(10).
				Background(bg).Border(1, border).TextColor(text).
				Transition(myui.ElementTransition{
					Duration:     200 * time.Millisecond,
					Enter:        &myui.Motion{Y: 12, Opacity: 0},
					Exit:         &myui.Motion{Y: 12, Opacity: 0},
					ExitDuration: 160 * time.Millisecond,
				})
			box.Children(func() {
				myui.Text(c, t.Title).FontSize(13).FontWeight(500).Grow(1)
			})
		}
	})
}
