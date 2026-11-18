package service

import (
	"context"
	"fmt"
	"strings"

	"autocarsystem/internal/config"
	"autocarsystem/internal/domain"
	"autocarsystem/internal/store"

	"github.com/egoist/mygo/plugins/sqlite"
)

type Service struct {
	store *store.Store
	paths config.Paths
}

func New(st *store.Store, paths config.Paths) *Service {
	return &Service{store: st, paths: paths}
}

func (s *Service) Bootstrap(ctx context.Context) error {
	rows, err := s.store.Query(ctx, "SELECT COUNT(*) FROM category")
	if err != nil {
		return fmt.Errorf("读取分类数量失败: %w", err)
	}
	if len(rows.Values) > 0 && len(rows.Values[0]) > 0 {
		if count, ok := rows.Values[0][0].(int64); ok && count > 0 {
			return nil
		}
	}

	defaults := []string{"手缝把套", "软包脚垫", "通用把套", "其他版型"}
	statements := make([]sqlite.Statement, 0, len(defaults))
	for i, name := range defaults {
		statements = append(statements, sqlite.Statement{
			SQL:  "INSERT INTO category(name, sort, enabled) VALUES (?, ?, 1)",
			Args: []any{name, i},
		})
	}
	if _, err := s.store.Transaction(ctx, statements); err != nil {
		return fmt.Errorf("初始化默认分类失败: %w", err)
	}
	return nil
}

func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := s.store.Query(ctx, "SELECT id, name, sort, enabled FROM category ORDER BY sort, id")
	if err != nil {
		return nil, fmt.Errorf("读取分类失败: %w", err)
	}

	categories := make([]domain.Category, 0, len(rows.Values))
	for _, values := range rows.Values {
		if len(values) < 4 {
			continue
		}
		categories = append(categories, domain.Category{
			ID:      valueInt64(values[0]),
			Name:    valueString(values[1]),
			Sort:    int(valueInt64(values[2])),
			Enabled: valueInt64(values[3]) != 0,
		})
	}
	return categories, nil
}

func (s *Service) CreateCategory(ctx context.Context, name string, sort int) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("分类名称不能为空")
	}
	_, err := s.store.Execute(ctx, "INSERT INTO category(name, sort, enabled) VALUES (?, ?, 1)", name, sort)
	if err != nil {
		return fmt.Errorf("新增分类失败: %w", err)
	}
	return nil
}

func (s *Service) SetCategoryEnabled(ctx context.Context, id int64, enabled bool) error {
	value := 0
	if enabled {
		value = 1
	}
	_, err := s.store.Execute(ctx, "UPDATE category SET enabled = ? WHERE id = ?", value, id)
	if err != nil {
		return fmt.Errorf("设置分类状态失败: %w", err)
	}
	return nil
}

func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	rows, err := s.store.Query(ctx, "SELECT COUNT(*) FROM youvision WHERE category_id = ?", id)
	if err != nil {
		return fmt.Errorf("检查分类引用失败: %w", err)
	}
	if len(rows.Values) > 0 && len(rows.Values[0]) > 0 && valueInt64(rows.Values[0][0]) > 0 {
		return fmt.Errorf("该分类下存在版型，禁止删除")
	}
	if _, err := s.store.Execute(ctx, "DELETE FROM category WHERE id = ?", id); err != nil {
		return fmt.Errorf("删除分类失败: %w", err)
	}
	return nil
}

func (s *Service) Counts(ctx context.Context) (map[string]int64, error) {
	tables := []string{"brand", "firm", "series", "spec", "category", "youvision", "gallery"}
	result := make(map[string]int64, len(tables))
	for _, table := range tables {
		rows, err := s.store.Query(ctx, "SELECT COUNT(*) FROM "+table)
		if err != nil {
			return nil, fmt.Errorf("统计 %s 失败: %w", table, err)
		}
		result[table] = 0
		if len(rows.Values) > 0 && len(rows.Values[0]) > 0 {
			result[table] = valueInt64(rows.Values[0][0])
		}
	}
	return result, nil
}

func valueInt64(v any) int64 {
	switch value := v.(type) {
	case int64:
		return value
	case float64:
		return int64(value)
	case bool:
		if value {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func valueString(v any) string {
	if value, ok := v.(string); ok {
		return value
	}
	return fmt.Sprint(v)
}
