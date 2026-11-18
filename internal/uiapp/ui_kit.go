package uiapp

import (
	"autocarsystem/internal/i18n"
	"fmt"
	"github.com/egoist/mygo"
	myui "github.com/egoist/mygo/ui"
	"strconv"
	"strings"
	"time"
)

func paletteFor(dark bool) palette {
	if dark {
		return palette{
			dark:     true,
			bg:       myui.Hex("#141416"),
			sidebar:  myui.Hex("#1c1c1f"),
			text:     myui.Hex("#f5f5f7"),
			muted:    myui.RGBA(235, 235, 245, 0.56),
			faint:    myui.RGBA(235, 235, 245, 0.3),
			hairline: myui.RGBA(255, 255, 255, 0.07),
			hover:    myui.RGBA(255, 255, 255, 0.06),
			press:    myui.RGBA(255, 255, 255, 0.1),
			accent:   myui.Oklch(0.66, 0.2, 22),
			onAccent: myui.Hex("#ffffff"),
		}
	}
	return palette{
		bg:       myui.Hex("#ffffff"),
		sidebar:  myui.Hex("#f4f4f6"),
		text:     myui.Hex("#1d1d1f"),
		muted:    myui.RGBA(60, 60, 67, 0.62),
		faint:    myui.RGBA(60, 60, 67, 0.32),
		hairline: myui.RGBA(0, 0, 0, 0.07),
		hover:    myui.RGBA(0, 0, 0, 0.04),
		press:    myui.RGBA(0, 0, 0, 0.08),
		accent:   myui.Oklch(0.6, 0.21, 24),
		onAccent: myui.Hex("#ffffff"),
	}
}

func sidebarGradient(dark bool) myui.LinearGradient {
	if dark {
		return myui.LinearGradient{Angle: 180, From: myui.Hex("#1c1c1f"), To: myui.Hex("#26262b")}
	}
	return myui.LinearGradient{Angle: 180, From: myui.Hex("#f4f4f6"), To: myui.Hex("#e9e9ef")}
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

func (a *App) statusLine() string {
	if strings.TrimSpace(a.status) != "" {
		return i18n.T("状态：") + a.status
	}
	return fmt.Sprintf(
		i18n.T("根目录：%s    图集：%d    版型：%d    分类：%d"),
		a.paths.Root,
		a.countInt("gallery"),
		a.countInt("youvision"),
		a.countInt("category"),
	)
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
