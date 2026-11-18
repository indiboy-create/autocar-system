// Package fileutil 提供跨模块复用的文件工具，避免 config/sync 之间产生导入环。
package fileutil

import (
	"os"
	"path/filepath"
)

// WriteFileAtomic 把 data 原子地写入 path：先在同目录写临时文件并 Sync，
// 再 rename 覆盖目标。崩溃或断电时目标文件要么是旧内容要么是新内容，
// 不会被截断成半截 JSON。Windows 上 os.Rename 不能覆盖已存在文件，
// 因此先把旧文件改名为 .old，失败时尽力回滚。
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-"+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	backup := tmpName + ".old"
	os.Remove(backup) // 忽略上一次崩溃残留
	if err := os.Rename(path, backup); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Rename(backup, path) // 尽力回滚
		return err
	}
	os.Remove(backup)
	return nil
}
