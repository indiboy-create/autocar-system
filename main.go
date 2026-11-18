package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"autocarsystem/internal/config"
	"autocarsystem/internal/service"
	"autocarsystem/internal/store"
	"autocarsystem/internal/uiapp"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

//go:embed resources/tray.png
var trayIcon []byte

func main() {
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}

	paths, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := paths.EnsureDirs(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	st, err := store.OpenAndMigrate(ctx, paths.DBPath, paths.BackupDir)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	svc := service.New(st, paths)
	if err := svc.Bootstrap(ctx); err != nil {
		log.Fatal(err)
	}

	app := uiapp.New(svc, paths)

	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:                "autoCarSystem",
			Width:                1120,
			Height:               760,
			MinWidth:             880,
			MinHeight:            600,
			TitleBarStyle:        mygo.TitleBarHidden,
			TitleBarHeight:       46,
			TrafficLightPosition: &mygo.Point{X: 20, Y: 20},
			AutoHideMenuBar:      true,
			Content:              ui.View(app.View),
		})
		app.AttachWindow(win)
		win.SetWindowButtonsTheme(mygo.ThemeSystem)

		mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) {
			app.SetQuitting(true)
		})
		win.OnClose(func(e *mygo.CloseEvent) {
			if !app.Quitting() {
				e.PreventDefault()
				win.Hide()
			}
		})

		tray, err := mygo.NewTray(mygo.TrayOptions{
			Icon:    trayIcon,
			ToolTip: "autoCarSystem",
			Menu: mygo.NewMenu([]*mygo.MenuItem{
				{
					Label: "显示 / 隐藏主窗口",
					Click: func(*mygo.MenuItem, *mygo.Window) {
						if win.IsVisible() {
							win.Hide()
							return
						}
						win.Show()
						win.Focus()
					},
				},
				mygo.Separator(),
				{Role: mygo.RoleQuit},
			}),
		})
		if err != nil {
			log.Printf("创建托盘失败: %v", err)
		} else {
			app.AttachTray(tray)
			tray.OnClick(win.Show)
		}

		mygo.App.OnSecondInstance(func([]string, string) {
			if !win.IsVisible() {
				win.Show()
			}
			win.Restore()
			win.Focus()
		})
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(fmt.Errorf("应用运行失败: %w", err))
	}
}
