package lib

import (
	"path/filepath"
	"runtime"
	"testing"
)

func getLocalConfigPath(file string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), file)
}

func Test_SaveConfig(t *testing.T) {

	err, conf := NewConfig(getLocalConfigPath("../config.json"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	err = conf.Save()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log("PASS")
}
