package lib

import (
	"os"
	"path/filepath"
	"time"
)

type Cleaner struct {
	tmpDir        string
	cleanupPeriod time.Duration // 清理间隔时间
	fileAgeLimit  time.Duration // 文件最大保存时间
	ticker        *time.Ticker
}

func NewCleaner(cleanupPeriod time.Duration, fileAgeLimit time.Duration) *Cleaner {
	return &Cleaner{
		tmpDir:        os.TempDir(),
		cleanupPeriod: cleanupPeriod,
		fileAgeLimit:  fileAgeLimit,
	}
}

func (c *Cleaner) Start() {
	c.ticker = time.NewTicker(c.cleanupPeriod)
	go func() {
		for range c.ticker.C {
			c.cleanTmpDir()
		}
	}()
}

func (c *Cleaner) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
}

func (c *Cleaner) cleanTmpDir() {
	Log.Infof("Cleaning tmp directory: %s\n", c.tmpDir)

	entries, err := os.ReadDir(c.tmpDir)
	if err != nil {
		Log.Errorf("Failed to read tmp directory: %v\n", err)
		return
	}

	now := time.Now()
	for _, entry := range entries {
		fileInfo, err := entry.Info()
		if err != nil {
			Log.Errorf("Failed to get file info for %s: %v\n", entry.Name(), err)
			continue
		}

		filePath := filepath.Join(c.tmpDir, entry.Name())
		if now.Sub(fileInfo.ModTime()) > c.fileAgeLimit {
			err := os.Remove(filePath)
			if err != nil {
				Log.Errorf("Failed to delete file %s: %v\n", filePath, err)
			} else {
				Log.Infof("Deleted file %s\n", filePath)
			}
		}
	}
}