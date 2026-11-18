package uiapp

import (
	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/media"
	"context"
	"fmt"
	"github.com/egoist/mygo"
	myui "github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (a *App) renderGallery(c *myui.Context) {
	myui.Text(c, i18n.T("图集列表")).FontSize(18).Bold()
	myui.Row(c).Gap(8).Children(func() {
		if myui.TextInput(c.Key("gal-yuvision"), &a.galYuvisionID).Placeholder(i18n.T("版型 ID")).Submitted() {
			a.createGallery()
		}
		myui.TextInput(c.Key("gal-title"), &a.galTitle).Placeholder(i18n.T("图集标题"))
		if myui.Button(c, i18n.T("新增图集")).Clicked() {
			a.createGallery()
		}
	})
	myui.Text(c, i18n.T("新增后请选择图片上传；封面自动取第一张图片。"))

	myui.Row(c).Gap(8).Children(func() {
		if myui.TextInput(c.Key("gal-search"), &a.gallerySearch).Placeholder(i18n.T("搜索标题 / 图集 ID / 版型 ID")).Changed() {
			a.Refresh()
		}
		if myui.Button(c, i18n.T("搜索")).Clicked() {
			a.Refresh()
		}
		if myui.Button(c, i18n.T("清除筛选")).Clicked() {
			a.gallerySearch = ""
			a.filterBrandID = ""
			a.filterFirmID = ""
			a.filterSeriesID = ""
			a.filterCategory = ""
			a.filterYuvision = ""
			a.Refresh()
		}
	})
	myui.Row(c).Gap(8).Children(func() {
		if myui.TextInput(c.Key("filter-brand"), &a.filterBrandID).Placeholder(i18n.T("品牌 ID")).Changed() {
			a.Refresh()
		}
		if myui.TextInput(c.Key("filter-firm"), &a.filterFirmID).Placeholder(i18n.T("厂商 ID")).Changed() {
			a.Refresh()
		}
		if myui.TextInput(c.Key("filter-series"), &a.filterSeriesID).Placeholder(i18n.T("车系 ID")).Changed() {
			a.Refresh()
		}
	})
	myui.Row(c).Gap(8).Children(func() {
		if myui.TextInput(c.Key("filter-category"), &a.filterCategory).Placeholder(i18n.T("分类 ID")).Changed() {
			a.Refresh()
		}
		if myui.TextInput(c.Key("filter-yuvision"), &a.filterYuvision).Placeholder(i18n.T("版型 ID")).Changed() {
			a.Refresh()
		}
	})
	myui.Row(c).Gap(8).Children(func() {
		if myui.Button(c, i18n.T("更新时间")).Clicked() {
			a.gallerySort = "updated_desc"
			a.Refresh()
		}
		if myui.Button(c, i18n.T("创建时间")).Clicked() {
			a.gallerySort = "created_desc"
			a.Refresh()
		}
		if myui.Button(c, i18n.T("标题")).Clicked() {
			a.gallerySort = "title_asc"
			a.Refresh()
		}
		if myui.Button(c, i18n.T("图片数")).Clicked() {
			a.gallerySort = "image_count_desc"
			a.Refresh()
		}
		if myui.Button(c, "ID").Clicked() {
			a.gallerySort = "id_desc"
			a.Refresh()
		}
	})

	if len(a.galleries) > 0 {
		myui.GridView(c, &a.gridState, len(a.galleries), 180, 150, func(i int) {
			item := a.galleries[i]
			myui.Column(c.Key(fmt.Sprintf("gallery-card-%d", item.ID))).Gap(6).Children(func() {
				path := a.galleryCoverPath(item)
				if bitmap := a.bitmapAt(path); bitmap != nil {
					myui.Image(c, bitmap).
						Size(168, 118).
						Fit(myui.Cover).
						Radius(8).
						OnClick(func() { a.openLightbox(item) })
				} else {
					myui.Text(c, i18n.T("暂无图片"))
				}
				myui.Text(c, fmt.Sprintf("#%d %s", item.ID, item.Title))
				if myui.Button(c, i18n.T("删除")).Clicked() {
					a.deleteGallery(item.ID)
				}
			})
		})
	} else {
		myui.Text(c, i18n.T("暂无图集。请先创建版型，再创建图集。"))
	}

	myui.Text(c, i18n.T("图集图片上传")).FontSize(16).Bold()
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("gal-upload-id"), &a.galUploadID).Placeholder(i18n.T("图集 ID"))
		myui.TextInput(c.Key("gal-upload-src"), &a.galUploadSrc).Placeholder(i18n.T("本地图片完整路径"))
		if myui.Button(c, i18n.T("选择图片")).Clicked() {
			idText := a.galUploadID
			go a.pickGalleryImages(idText)
		}
		if myui.Button(c, i18n.T("上传路径")).Clicked() {
			a.uploadGalleryImage()
		}
	})
}

func (a *App) renderYuvision(c *myui.Context) {
	myui.Text(c, i18n.T("版型管理")).FontSize(18).Bold()
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("yv-brand"), &a.yvBrandID).Placeholder(i18n.T("品牌 ID"))
		myui.TextInput(c.Key("yv-firm"), &a.yvFirmID).Placeholder(i18n.T("厂商 ID"))
		myui.TextInput(c.Key("yv-series"), &a.yvSeriesID).Placeholder(i18n.T("车系 ID"))
		myui.TextInput(c.Key("yv-category"), &a.yvCategory).Placeholder(i18n.T("分类 ID"))
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("yv-name"), &a.yvName).Placeholder(i18n.T("版型名称"))
		myui.TextInput(c.Key("yv-desc"), &a.yvDesc).Placeholder(i18n.T("描述"))
		if myui.Button(c, i18n.T("新增版型")).Clicked() {
			a.createYuvision()
		}
	})
	myui.Text(c, i18n.T("新增版型时不允许选择禁用分类；车系、品牌、厂商必须匹配。"))

	yuvisions := make([]domain.Yuvision, 0, len(a.yuvisions))
	for _, item := range a.yuvisions {
		if a.yuvFilterCategory == 0 || item.CategoryID == a.yuvFilterCategory {
			yuvisions = append(yuvisions, item)
		}
	}
	for _, item := range yuvisions {
		current := item
		myui.Row(c.Key(fmt.Sprintf("youvision-%d", current.ID))).Gap(8).Children(func() {
			myui.Text(c, fmt.Sprintf(i18n.T("版型 #%d  %s  分类 #%d"), current.ID, current.Name, current.CategoryID)).Grow(1)
			if myui.Button(c, i18n.T("删除")).Clicked() {
				a.deleteYuvision(current.ID)
			}
		})
	}

	myui.Text(c, i18n.T("版型图片上传")).FontSize(16).Bold()
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("yv-upload-id"), &a.yvUploadID).Placeholder(i18n.T("版型 ID"))
		myui.TextInput(c.Key("yv-upload-src"), &a.yvUploadSrc).Placeholder(i18n.T("本地图片完整路径"))
		if myui.Button(c, i18n.T("选择图片")).Clicked() {
			idText := a.yvUploadID
			go a.pickYuvisionImages(idText)
		}
		if myui.Button(c, i18n.T("上传路径")).Clicked() {
			a.uploadYuvisionImage()
		}
	})
	if a.countInt("youvision") == 0 {
		myui.Text(c, i18n.T("暂无版型。请先同步目录数据或手工准备品牌、厂商、车系记录。"))
	}
}

func (a *App) renderCategories(c *myui.Context) {
	myui.Row(c).Gap(8).Children(func() {
		if myui.TextInput(c, &a.newName).Placeholder(i18n.T("新分类名称")).Submitted() {
			a.createCategory()
		}
		if myui.Button(c, i18n.T("新增分类")).Clicked() {
			a.createCategory()
		}
	})
	for _, category := range a.categories {
		item := category
		myui.Row(c.Key(item.ID)).Gap(8).Children(func() {
			label := item.Name
			if !item.Enabled {
				label += i18n.T("（已禁用）")
			}
			myui.Text(c, label).Grow(1)
			action := i18n.T("禁用")
			if !item.Enabled {
				action = i18n.T("启用")
			}
			if myui.Button(c, action).Clicked() {
				err := a.svc.SetCategoryEnabled(context.Background(), item.ID, !item.Enabled)
				if err != nil {
					a.setStatus(err.Error())
				} else {
					a.setStatus("")
					a.Refresh()
				}
			}
			if myui.Button(c, i18n.T("删除")).Clicked() {
				if err := a.svc.DeleteCategory(context.Background(), item.ID); err != nil {
					a.setStatus(err.Error())
				} else {
					a.setStatus("")
					a.Refresh()
				}
			}
		})
	}
}

func (a *App) createCategory() {
	name := strings.TrimSpace(a.newName)
	if name == "" {
		a.setStatus(i18n.T("分类名称不能为空"))
		return
	}
	if err := a.svc.CreateCategory(context.Background(), name, len(a.categories)); err != nil {
		a.setStatus(err.Error())
		return
	}
	a.newName = ""
	a.setStatus("")
	a.Refresh()
}

func (a *App) createYuvision() {
	ctx := context.Background()
	brandID, err := parseID(a.yvBrandID)
	if err != nil {
		a.setStatus(i18n.T("品牌 ID 无效"))
		return
	}
	firmID, err := parseID(a.yvFirmID)
	if err != nil {
		a.setStatus(i18n.T("厂商 ID 无效"))
		return
	}
	seriesID, err := parseID(a.yvSeriesID)
	if err != nil {
		a.setStatus(i18n.T("车系 ID 无效"))
		return
	}
	categoryID, err := parseID(a.yvCategory)
	if err != nil {
		a.setStatus(i18n.T("分类 ID 无效"))
		return
	}

	id, err := a.svc.CreateYuvision(ctx, brandID, firmID, seriesID, categoryID, a.yvName, a.yvDesc)
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	a.yvName, a.yvDesc = "", ""
	a.yvUploadID = strconv.FormatInt(id, 10)
	a.setStatus(fmt.Sprintf(i18n.T("版型 #%d 已创建，请选择图片上传"), id))
	a.Refresh()
}

func (a *App) deleteYuvision(id int64) {
	if err := a.svc.DeleteYuvision(context.Background(), id); err != nil {
		a.setStatus(err.Error())
		return
	}
	a.setStatus("")
	a.Refresh()
}

func (a *App) uploadYuvisionImage() {
	id, err := parseID(a.yvUploadID)
	if err != nil {
		a.setStatus(i18n.T("版型 ID 无效"))
		return
	}
	rel, err := a.svc.AddYuvisionImage(context.Background(), id, strings.TrimSpace(a.yvUploadSrc))
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	a.setStatus(i18n.T("已上传图片：") + rel)
	a.Refresh()
}

func (a *App) createGallery() {
	yuvisionID, err := parseID(a.galYuvisionID)
	if err != nil {
		a.setStatus(i18n.T("版型 ID 无效"))
		return
	}
	id, err := a.svc.CreateGallery(context.Background(), yuvisionID, a.galTitle)
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	a.galTitle = ""
	a.galUploadID = strconv.FormatInt(id, 10)
	a.setStatus(fmt.Sprintf(i18n.T("图集 #%d 已创建，请选择图片上传"), id))
	a.Refresh()
}

func (a *App) deleteGallery(id int64) {
	if err := a.svc.DeleteGallery(context.Background(), id); err != nil {
		a.setStatus(err.Error())
		return
	}
	a.setStatus("")
	a.Refresh()
}

func (a *App) uploadGalleryImage() {
	id, err := parseID(a.galUploadID)
	if err != nil {
		a.setStatus(i18n.T("图集 ID 无效"))
		return
	}
	rel, err := a.svc.AddGalleryImage(context.Background(), id, strings.TrimSpace(a.galUploadSrc))
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	a.setStatus(i18n.T("已上传图片：") + rel)
	a.Refresh()
}

func (a *App) pickYuvisionImages(idText string) {
	paths, err := a.openImageDialog(i18n.T("选择版型图片"))
	if err != nil {
		a.uiStatus(err.Error())
		return
	}
	if paths == nil {
		return
	}
	id, err := parseID(idText)
	if err != nil {
		a.uiStatus(i18n.T("版型 ID 无效"))
		return
	}
	added := 0
	for _, path := range paths {
		if _, err := a.svc.AddYuvisionImage(context.Background(), id, path); err != nil {
			a.uiStatus(err.Error())
			return
		}
		added++
	}
	a.uiStatus(fmt.Sprintf(i18n.T("已上传 %d 张版型图片"), added))
}

func (a *App) pickGalleryImages(idText string) {
	paths, err := a.openImageDialog(i18n.T("选择图集图片"))
	if err != nil {
		a.uiStatus(err.Error())
		return
	}
	if paths == nil {
		return
	}
	id, err := parseID(idText)
	if err != nil {
		a.uiStatus(i18n.T("图集 ID 无效"))
		return
	}
	added := 0
	for _, path := range paths {
		if _, err := a.svc.AddGalleryImage(context.Background(), id, path); err != nil {
			a.uiStatus(err.Error())
			return
		}
		added++
	}
	a.uiStatus(fmt.Sprintf(i18n.T("已上传 %d 张图集图片"), added))
}

func (a *App) openImageDialog(title string) ([]string, error) {
	return mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: a.win,
		Title:  title,
		Filters: []mygo.FileFilter{{
			Name:       i18n.T("图片"),
			Extensions: []string{"jpg", "jpeg", "png", "gif", "webp"},
		}},
		Multiple: true,
	})
}

func (a *App) renderLightbox(c *myui.Context) {
	myui.Column(c).Fill().Center().Padding(20).Gap(12).Children(func() {
		myui.Text(c, a.lightboxTitle).FontSize(18).Bold()
		if a.lightboxImage != nil {
			myui.Image(c, a.lightboxImage).Size(900, 620).Fit(myui.Contain).Radius(8)
		} else {
			myui.Text(c, i18n.T("图片无法预览"))
		}
		if myui.Button(c, i18n.T("关闭")).Clicked() {
			if a.lightboxWin != nil {
				a.lightboxWin.Hide()
			}
		}
	})
}

func (a *App) openLightbox(item domain.Gallery) {
	path := a.galleryOriginalPath(item)
	if path == "" {
		a.uiStatus(i18n.T("该图集暂无图片"))
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		a.uiStatus(err.Error())
		return
	}
	bitmap, err := myui.DecodeBitmap(data)
	if err != nil {
		a.uiStatus(err.Error())
		return
	}
	a.lightboxImage = bitmap
	a.lightboxTitle = fmt.Sprintf(i18n.T("图集 #%d %s"), item.ID, item.Title)

	if a.lightboxWin == nil {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:           i18n.T("图片预览"),
			Width:           960,
			Height:          720,
			AutoHideMenuBar: true,
			Content:         myui.View(a.renderLightbox),
		})
		win.OnClose(func(e *mygo.CloseEvent) {
			if !a.quitting {
				e.PreventDefault()
				win.Hide()
			}
		})
		a.lightboxWin = win
	}
	a.lightboxWin.Show()
	a.lightboxWin.Focus()
}

func (a *App) bitmapAt(path string) *myui.Bitmap {
	if path == "" {
		return nil
	}
	return a.coverBitmaps[path]
}

func (a *App) galleryCoverPath(item domain.Gallery) string {
	if item.CoverImage == "" {
		return ""
	}
	base := filepath.Join(a.paths.Gallery, fmt.Sprint(item.SeriesID), fmt.Sprint(item.ID))
	thumb := filepath.Join(base, filepath.FromSlash(media.ThumbnailRel(item.CoverImage)))
	if _, err := os.Stat(thumb); err == nil {
		return thumb
	}
	return filepath.Join(base, filepath.FromSlash(item.CoverImage))
}

func (a *App) galleryOriginalPath(item domain.Gallery) string {
	if item.CoverImage == "" {
		return ""
	}
	base := filepath.Join(a.paths.Gallery, fmt.Sprint(item.SeriesID), fmt.Sprint(item.ID))
	return filepath.Join(base, filepath.FromSlash(item.CoverImage))
}

func (a *App) countInt(name string) int64 {
	if a.counts == nil {
		return 0
	}
	return a.counts[name]
}

func (a *App) galleryFilter() domain.GalleryFilter {
	filter := domain.GalleryFilter{
		Keyword: a.gallerySearch,
		Sort:    a.gallerySort,
	}

	fields := []struct {
		name  string
		value string
		out   *int64
	}{
		{i18n.T("品牌 ID"), a.filterBrandID, &filter.BrandID},
		{i18n.T("厂商 ID"), a.filterFirmID, &filter.FirmID},
		{i18n.T("车系 ID"), a.filterSeriesID, &filter.SeriesID},
		{i18n.T("分类 ID"), a.filterCategory, &filter.CategoryID},
		{i18n.T("版型 ID"), a.filterYuvision, &filter.YouvisionID},
	}
	for _, field := range fields {
		value, err := parseOptionalID(field.value)
		if err != nil {
			a.setStatus(field.name + i18n.T(" 无效"))
			continue
		}
		*field.out = value
	}
	return filter
}
