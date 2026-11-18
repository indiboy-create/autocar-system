package sync

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type StageStats struct {
	Processed int64 `json:"processed"`
	Inserted  int64 `json:"inserted"`
	Skipped   int64 `json:"skipped"`
	Failed    int64 `json:"failed"`
}

type State struct {
	LastSyncAt string                `json:"last_sync_at"`
	Stage      string                `json:"stage"`
	Mode       string                `json:"mode"`
	Stats      map[string]StageStats `json:"stats"`
}

type Conflict struct {
	Entity    string `json:"entity"`
	ID        string `json:"id"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

func NewState() State {
	return State{
		Stage: "idle",
		Mode:  "full_scan",
		Stats: map[string]StageStats{
			"brand":  {},
			"firm":   {},
			"series": {},
			"spec":   {},
		},
	}
}

func LoadState(path string) (State, error) {
	state := NewState()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return state, fmt.Errorf("读取同步状态失败: %w", err)
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("解析同步状态失败: %w", err)
	}
	if state.Stats == nil {
		state.Stats = NewState().Stats
	}
	if state.Mode == "" {
		state.Mode = "full_scan"
	}
	return state, nil
}

func SaveState(path string, state State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建同步状态目录失败: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化同步状态失败: %w", err)
	}
	if err := writeFileAtomic(path, data, 0o644); err != nil {
		return fmt.Errorf("写入同步状态失败: %w", err)
	}
	return nil
}

// writeFileAtomic 把数据先写进同目录的临时文件，再替换目标文件，避免进程在
// 写入中途崩溃时留下被截断的目标文件。Windows 上 os.Rename 不能覆盖已存在的
// 文件，因此 rename 前先移走原文件；若 rename 失败则尽力把原文件恢复回去。
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
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
		// 尽力回滚：把旧文件放回原路径
		os.Rename(backup, path)
		return err
	}
	os.Remove(backup)
	return nil
}

func AppendConflict(path string, conflict Conflict) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建冲突日志目录失败: %w", err)
	}
	line, err := json.Marshal(conflict)
	if err != nil {
		return fmt.Errorf("序列化冲突日志失败: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("打开冲突日志失败: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if _, err := w.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("写入冲突日志失败: %w", err)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("刷新冲突日志失败: %w", err)
	}
	return nil
}
