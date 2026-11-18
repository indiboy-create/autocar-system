package uiapp

import (
	"context"
	"os"

	"autocarsystem/internal/domain"

	myui "github.com/egoist/mygo/ui"
)

// refreshResult 是一次后台刷新的全部结果。字段只在后台 goroutine 写入，
// 经 Window.Update 回投到主线程后才被应用到 App 状态。
type refreshResult struct {
	categories []domain.Category
	yuvisions  []domain.Yuvision
	galleries  []domain.Gallery
	counts     map[string]int64
	brands     []domain.Brand
	firms      []domain.Firm
	series     []domain.Series
	specs      []domain.Spec
	backups    []domain.BackupInfo
	covers     map[string]*myui.Bitmap
	withCovers bool // 本次是否按需解码了封面（决定 apply 时是否覆盖缓存）
	errors     []string
}

// fetchRefresh 在后台执行所有数据库查询与封面解码，不触碰 App 的可变状态。
// withCovers 为 false 时跳过封面解码：非图集页不需要封面，省去最多 200 次
// 磁盘读取与图片解码。
func (a *App) fetchRefresh(ctx context.Context, filter domain.GalleryFilter, withCovers bool) refreshResult {
	var r refreshResult

	if v, err := a.svc.ListCategories(ctx); err != nil {
		r.errors = append(r.errors, err.Error())
	} else {
		r.categories = v
	}

	if v, err := a.svc.ListYuvisions(ctx, 200); err != nil {
		r.errors = append(r.errors, err.Error())
	} else {
		r.yuvisions = v
	}

	if v, err := a.svc.SearchGalleries(ctx, filter, 200); err != nil {
		r.errors = append(r.errors, err.Error())
	} else {
		r.galleries = v
		if withCovers {
			r.covers = a.buildCoverBitmaps(v)
			r.withCovers = true
		}
	}

	if v, err := a.svc.Counts(ctx); err != nil {
		r.errors = append(r.errors, err.Error())
	} else {
		r.counts = v
	}

	if v, err := a.svc.ListBrands(ctx, 200); err == nil {
		r.brands = v
	}
	if v, err := a.svc.ListFirms(ctx, 200); err == nil {
		r.firms = v
	}
	if v, err := a.svc.ListSeries(ctx, 200); err == nil {
		r.series = v
	}
	if v, err := a.svc.ListSpecs(ctx, 200); err == nil {
		r.specs = v
	}
	if v, err := a.svc.ListBackups(); err == nil {
		r.backups = v
	}

	return r
}

// applyRefresh 在主线程应用后台结果。只有查询成功（对应字段非 nil）时才覆盖
// 现有数据，保持与同步版本一致：失败只上报状态，不清空旧数据。
func (a *App) applyRefresh(r refreshResult, seq uint64, async bool) {
	for _, message := range r.errors {
		a.setStatus(message)
	}
	if r.categories != nil {
		a.categories = r.categories
	}
	if r.yuvisions != nil {
		a.yuvisions = r.yuvisions
	}
	if r.galleries != nil {
		a.galleries = r.galleries
	}
	if r.withCovers {
		a.coverBitmaps = r.covers
	}
	if r.counts != nil {
		a.counts = r.counts
	}
	if r.brands != nil {
		a.brands = r.brands
	}
	if r.firms != nil {
		a.firms = r.firms
	}
	if r.series != nil {
		a.series = r.series
	}
	if r.specs != nil {
		a.specs = r.specs
	}
	if r.backups != nil {
		a.backups = r.backups
	}
	if async {
		a.invalidateWindows()
	}
}

// buildCoverBitmaps 根据图集列表读取并解码封面。只读 a.paths（启动后不变），
// 因此可安全地在后台 goroutine 调用，返回的 map 由调用方回投到主线程。
func (a *App) buildCoverBitmaps(galleries []domain.Gallery) map[string]*myui.Bitmap {
	bitmaps := make(map[string]*myui.Bitmap)
	for _, item := range galleries {
		path := a.galleryCoverPath(item)
		if path == "" || bitmaps[path] != nil {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		bitmap, err := myui.DecodeBitmap(data)
		if err != nil {
			continue
		}
		bitmaps[path] = bitmap
	}
	return bitmaps
}
