package store

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"autocarsystem/internal/i18n"
	"github.com/egoist/mygo/plugins/sqlite"
)

type Store struct {
	db      *sqlite.DB
	path    string
	writeMu sync.Mutex
}

// EnsureSeeded 在 dbPath 不存在时，把编译进程序的预置库原子地释放到该位置。
// 已存在则原样跳过，绝不覆盖用户在运行中积累的数据。
func EnsureSeeded(dbPath string, seed []byte) error {
	if _, err := os.Stat(dbPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return i18n.Errorf("检查预置数据库失败: %w", err)
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return i18n.Errorf("创建数据库目录失败: %w", err)
	}
	if len(seed) == 0 {
		return i18n.Errorf("预置数据库内容为空")
	}

	tmp, err := os.CreateTemp(dir, ".car-seed-*.db")
	if err != nil {
		return i18n.Errorf("创建临时预置库失败: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(seed); err != nil {
		_ = tmp.Close()
		return i18n.Errorf("写入预置库失败: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return i18n.Errorf("同步预置库失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return i18n.Errorf("关闭预置库失败: %w", err)
	}
	if err := os.Rename(tmpName, dbPath); err != nil {
		return i18n.Errorf("安装预置库失败: %w", err)
	}
	return nil
}

// Open 打开已存在的数据库，不执行建表或迁移。
// 预置库自带完整的表结构与基础数据；缺失时应由 EnsureSeeded 释放，
// 而不是在这里创建一个空库。
func Open(ctx context.Context, dbPath string) (*Store, error) {
	info, err := os.Stat(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, i18n.Errorf("数据库不存在: %s", dbPath)
		}
		return nil, i18n.Errorf("检查数据库失败: %w", err)
	}
	if info.Size() == 0 {
		return nil, i18n.Errorf("数据库文件为空: %s", dbPath)
	}

	db, err := openDB(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	return &Store{db: db, path: dbPath}, nil
}

// OpenAndMigrate 供测试构造临时空库使用；生产路径应使用 Open + EnsureSeeded。
func OpenAndMigrate(ctx context.Context, dbPath string, backupDir string) (*Store, error) {
	db, err := openDB(ctx, dbPath)
	if err != nil {
		return nil, err
	}

	version, err := readSchemaVersion(ctx, db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	if version < SchemaVersion {
		if err := db.Close(); err != nil {
			return nil, err
		}
		if err := backupBeforeMigration(dbPath, backupDir); err != nil {
			return nil, err
		}
		db, err = openDB(ctx, dbPath)
		if err != nil {
			return nil, err
		}
		if err := applySchema(ctx, db); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	return &Store{db: db, path: dbPath}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Path() string {
	return s.path
}

// Backup 生成一份一致性备份：先把 WAL 中的事务检查点写入主库，
// 再在写锁内复制数据库文件，避免复制过程中有写入导致的不一致。
func (s *Store) Backup(ctx context.Context, destPath string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	if _, err := s.db.Execute(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return i18n.Errorf("刷新数据库检查点失败: %w", err)
	}

	src, err := os.Open(s.path)
	if err != nil {
		return i18n.Errorf("打开数据库备份源失败: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return i18n.Errorf("创建数据库备份失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return i18n.Errorf("写入数据库备份失败: %w", err)
	}
	return nil
}

func (s *Store) Query(ctx context.Context, sql string, args ...any) (sqlite.Rows, error) {
	return s.db.Query(ctx, sql, args...)
}

func (s *Store) Execute(ctx context.Context, sql string, args ...any) (sqlite.Result, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.db.Execute(ctx, sql, args...)
}

func (s *Store) Transaction(ctx context.Context, statements []sqlite.Statement) ([]sqlite.Result, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.db.Transaction(ctx, statements)
}

func openDB(ctx context.Context, dbPath string) (*sqlite.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, i18n.Errorf("创建数据库目录失败: %w", err)
	}

	db, err := sqlite.Open(ctx, dbPath, sqlite.OpenOptions{BusyTimeout: 5 * time.Second})
	if err != nil {
		return nil, i18n.Errorf("打开数据库失败: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA foreign_keys = ON",
		"PRAGMA temp_store = MEMORY",
	}
	for _, pragma := range pragmas {
		if _, err := db.Execute(ctx, pragma); err != nil {
			_ = db.Close()
			return nil, i18n.Errorf("设置 SQLite 参数失败 %s: %w", pragma, err)
		}
	}

	return db, nil
}

func readSchemaVersion(ctx context.Context, db *sqlite.DB) (int64, error) {
	rows, err := db.Query(ctx, "PRAGMA user_version")
	if err != nil {
		return 0, i18n.Errorf("读取数据库版本失败: %w", err)
	}
	if len(rows.Values) == 0 || len(rows.Values[0]) == 0 {
		return 0, nil
	}
	switch v := rows.Values[0][0].(type) {
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	default:
		return 0, i18n.Errorf("数据库版本类型异常: %T", rows.Values[0][0])
	}
}

func applySchema(ctx context.Context, db *sqlite.DB) error {
	statements := make([]sqlite.Statement, 0, len(schemaStatements())+1)
	for _, sql := range schemaStatements() {
		statements = append(statements, sqlite.Statement{SQL: sql})
	}
	statements = append(statements, sqlite.Statement{SQL: fmt.Sprintf("PRAGMA user_version = %d", SchemaVersion)})
	if _, err := db.Transaction(ctx, statements); err != nil {
		return i18n.Errorf("执行数据库迁移失败: %w", err)
	}
	return nil
}

func backupBeforeMigration(dbPath string, backupDir string) error {
	info, err := os.Stat(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return i18n.Errorf("检查数据库文件失败: %w", err)
	}
	if info.Size() == 0 {
		return nil
	}

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return i18n.Errorf("创建备份目录失败: %w", err)
	}

	name := fmt.Sprintf("car.db.bak-%s", time.Now().Format("20060102-150405"))
	backupPath := filepath.Join(backupDir, name)
	for i := 2; fileExists(backupPath); i++ {
		backupPath = filepath.Join(backupDir, fmt.Sprintf("%s-%d", name, i))
	}

	src, err := os.Open(dbPath)
	if err != nil {
		return i18n.Errorf("打开数据库备份源失败: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(backupPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return i18n.Errorf("创建数据库备份失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return i18n.Errorf("写入数据库备份失败: %w", err)
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func splitSQL(sql string) []string {
	parts := strings.Split(sql, ";")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
