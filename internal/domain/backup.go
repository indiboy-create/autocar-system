package domain

import "time"

type BackupInfo struct {
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
}
