package lib

import (
	"html2pdf/tests"
	"testing"
)

func loadConfig() (*Config, error) {
	err, conf := NewConfig(tests.GetLocalPath("../config.json"))
	if err != nil {
		return nil, err
	}
	conf = conf.LoadWithENV()
	return conf, nil
}

func Test_SaveConfig(t *testing.T) {

	conf, err := loadConfig()
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
