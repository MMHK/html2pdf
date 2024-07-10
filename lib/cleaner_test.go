package lib

import (
	"testing"
	"time"
)

func TestCleaner_Start(t *testing.T) {
	cleaner := NewCleaner(time.Second*10, time.Hour * 24 * 15)

	defer cleaner.Stop()

	cleaner.Start()

	time.Sleep(time.Second *50)

	t.Log("PASS")
}