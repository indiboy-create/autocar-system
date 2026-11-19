package uiapp

import (
	"autocarsystem/internal/i18n"
	"fmt"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/glass"
	myui "github.com/egoist/mygo/ui"
	"strconv"
	"strings"
	"time"
)

// colorScheme 是一种配色方案：明暗各一套强调色，hue 决定左侧栏背景的色调。
// swatch 用于设置页的色块预览。
type colorScheme struct {
	key         string
	label       string // i18n 键
	swatch      myui.Color
	darkAccent  myui.Color
	lightAccent myui.Color
	hue         float32 // 侧栏背景色的 Oklch 色相
}

// colorSchemes 是内置的配色方案。Oklch 的色相决定颜色，明度/彩度按明暗主题分别调校，
// 保证强调色在两种背景下都有足够对比度；hue 让左侧栏带同色系的低饱和背景。
var colorSchemes = []colorScheme{
	{"orange", i18n.N("暖橙"), myui.Oklch(0.66, 0.2, 22), myui.Oklch(0.66, 0.2, 22), myui.Oklch(0.6, 0.21, 24), 22},
	{"rose", i18n.N("玫红"), myui.Oklch(0.68, 0.18, 355), myui.Oklch(0.68, 0.18, 355), myui.Oklch(0.6, 0.21, 355), 355},
	{"violet", i18n.N("紫罗兰"), myui.Oklch(0.68, 0.17, 300), myui.Oklch(0.68, 0.17, 300), myui.Oklch(0.58, 0.22, 300), 300},
	{"blue", i18n.N("海蓝"), myui.Oklch(0.68, 0.16, 250), myui.Oklch(0.68, 0.16, 250), myui.Oklch(0.58, 0.2, 255), 250},
	{"teal", i18n.N("青碧"), myui.Oklch(0.7, 0.13, 195), myui.Oklch(0.7, 0.13, 195), myui.Oklch(0.58, 0.14, 195), 195},
	{"green", i18n.N("松绿"), myui.Oklch(0.7, 0.15, 150), myui.Oklch(0.7, 0.15, 150), myui.Oklch(0.58, 0.16, 150), 150},
}

// sidebarColors 按方案色相派生左侧栏渐变的起止色：低彩度，既能看出色调又不喧宾夺主。
func (s colorScheme) sidebarColors(dark bool) (from, to myui.Color) {
	if dark {
		return myui.Oklch(0.215, 0.02, s.hue), myui.Oklch(0.27, 0.027, s.hue)
	}
	return myui.Oklch(0.965, 0.009, s.hue), myui.Oklch(0.925, 0.015, s.hue)
}

// currentColorSchemeKey 为当前生效的配色方案，由偏好设置在启动时初始化、在设置页切换时更新。
var currentColorSchemeKey = colorSchemes[0].key

func setColorScheme(key string) {
	for _, scheme := range colorSchemes {
		if scheme.key == key {
			currentColorSchemeKey = key
			return
		}
	}
	currentColorSchemeKey = colorSchemes[0].key
}

func activeColorScheme() colorScheme {
	for _, scheme := range colorSchemes {
		if scheme.key == currentColorSchemeKey {
			return scheme
		}
	}
	return colorSchemes[0]
}

func paletteFor(dark bool) palette {
	scheme := activeColorScheme()
	accent := scheme.lightAccent
	sidebarFrom, _ := scheme.sidebarColors(dark)
	if dark {
		accent = scheme.darkAccent
		return palette{
			dark:     true,
			bg:       myui.Hex("#141416"),
			sidebar:  sidebarFrom,
			text:     myui.Hex("#f5f5f7"),
			muted:    myui.RGBA(235, 235, 245, 0.56),
			faint:    myui.RGBA(235, 235, 245, 0.3),
			hairline: myui.RGBA(255, 255, 255, 0.07),
			hover:    myui.RGBA(255, 255, 255, 0.06),
			press:    myui.RGBA(255, 255, 255, 0.1),
			accent:   accent,
			onAccent: myui.Hex("#ffffff"),
		}
	}
	return palette{
		bg:       myui.Hex("#ffffff"),
		sidebar:  sidebarFrom,
		text:     myui.Hex("#1d1d1f"),
		muted:    myui.RGBA(60, 60, 67, 0.62),
		faint:    myui.RGBA(60, 60, 67, 0.32),
		hairline: myui.RGBA(0, 0, 0, 0.07),
		hover:    myui.RGBA(0, 0, 0, 0.04),
		press:    myui.RGBA(0, 0, 0, 0.08),
		accent:   accent,
		onAccent: myui.Hex("#ffffff"),
	}
}

func sidebarGradient(dark bool) myui.LinearGradient {
	from, to := activeColorScheme().sidebarColors(dark)
	return myui.LinearGradient{Angle: 180, From: from, To: to}
}

func navIconButton(c *myui.Context, key string, icon *myui.SVG, label string) myui.Element {
	p := paletteFor(c.Theme().Dark)
	b := myui.ButtonBase(c.Key(key)).Size(34, 34).Center().Radius(999).Label(label).
		TextColor(p.muted).FocusRing(false)
	bg := myui.Transparent
	if b.Hovered() {
		bg = myui.RGBA(128, 128, 128, 0.14)
	}
	b.Background(b.AnimateColor("bg", bg, myui.Animation{Duration: 120 * time.Millisecond}))
	b.Children(func() {
		myui.Icon(c, icon).Size(18, 18).TextColor(p.muted)
	})
	return b
}

// backToTopButton 在四个目录页下拉后浮现在右下角，点击回到列表顶部后自动消失。
// 用 Visible() 的可见首行判断是否已离开顶部；列表由各自的 ListState 滚动，
// 这里只读取上一帧状态，不改变布局。
func (a *App) backToTopButton(c *myui.Context, p palette, bottom float32) {
	if a.page != "catalog" {
		return
	}
	var list *myui.ListState
	switch a.catalogEntity {
	case "brand":
		list = &a.brandList
	case "firm":
		list = &a.firmList
	case "series":
		list = &a.seriesList
	case "spec":
		list = &a.specList
	default:
		return
	}
	first, _ := list.Visible()
	if first <= 0 {
		return
	}

	b := myui.ButtonBase(c.Key("back-to-top")).
		Absolute().Right(22).Bottom(bottom+12).Size(44, 44).Radius(999).Center().
		Cursor(myui.CursorPointer).
		Label(i18n.T("返回顶部")).FocusRing(false)
	bg := p.accent
	if b.Hovered() {
		bg = p.accent.Alpha(0.85)
	}
	b.Background(bg)
	b.Children(func() {
		myui.Icon(c, iconArrowUp).Size(20, 20).TextColor(p.onAccent)
	})
	if b.Clicked() {
		list.ScrollTo(0, myui.Start)
	}
}

func edgeShown(edge myui.Element, under bool) float32 {
	target := float32(0)
	if under {
		target = 1
	}
	return edge.AnimateTo("shown", target, myui.Animation{Duration: 180 * time.Millisecond})
}

func strokeIcon(paths string) *myui.SVG {
	return myui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">` + paths + `</svg>`))
}

func applyAppearance(appearance string) {
	src := mygo.ThemeSystem
	switch appearance {
	case "light":
		src = mygo.ThemeLight
	case "dark":
		src = mygo.ThemeDark
	}
	mygo.Theme.SetSource(src)
}

func (a *App) applyTheme(c *myui.Context) {
	t := *c.Theme()
	p := paletteFor(t.Dark)
	black, white := myui.Hex("#000000"), myui.Hex("#ffffff")
	t.Accent = p.accent
	t.AccentText = p.onAccent
	t.AccentHover = p.accent.MixOklab(black, 0.1)
	t.AccentPressed = p.accent.MixOklab(black, 0.2)
	if p.dark {
		t.AccentHover = p.accent.MixOklab(white, 0.15)
		t.AccentPressed = p.accent.MixOklab(black, 0.15)
	}
	c.SetTheme(&t)
}

// bottomBarHeight 是悬浮在主内容区底部的底栏高度，参考 mygo-music 的播放栏。
const bottomBarHeight float32 = 76

// statusBar 只保留底部栏的框架：毛玻璃边缘与空白底栏，不填充任何内容。
func (a *App) statusBar(c *myui.Context, p palette) {
	// 悬浮在内容之上：内容从下方滚过，毛玻璃边缘让栏与内容分层；
	// 静止时（栏下没有内容）边缘淡出，与 AppKit 的硬滚动边缘一致。
	edge := myui.Box(c.Key("bottom-edge")).Absolute().Left(0).Right(0).Bottom(0).Height(bottomBarHeight).
		Material(glass.ScrollEdge{Hard: true, Bottom: true, Background: p.bg})
	edge.Opacity(edgeShown(edge, a.underBottom))

	myui.Row(c.Key("status-bar")).Absolute().Left(0).Right(0).Bottom(0).Height(bottomBarHeight)
}

func (a *App) setStatus(message string) {
	a.status = strings.TrimSpace(i18n.T(message))
}

func (a *App) uiStatus(message string) {
	if a.win != nil {
		a.win.Update(func() {
			a.setStatus(message)
			a.Refresh()
		})
		return
	}
	a.setStatus(message)
	a.Refresh()
}

func parseOptionalID(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func parseID(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty id")
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func parseInt64Value(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

func parseFloatValue(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

func formatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
