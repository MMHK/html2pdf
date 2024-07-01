package lib

import "testing"

func TestHTTPService_Start(t *testing.T) {
	conf, err := loadConfig()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	http := NewHTTP(conf)
	http.Start()
}
