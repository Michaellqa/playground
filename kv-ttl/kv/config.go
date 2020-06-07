package kv

import "time"

type Configuration struct {
	BackupInterval time.Duration
	FileName       string
}

type configuration struct {
	backupInterval time.Duration
	fileName       string
}
