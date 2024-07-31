package lib

import (
	_ "html2pdf/tests"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewHTTP(t *testing.T) {
	conf := &Config{}

	conf.Listen = os.Getenv("LISTEN")
	conf.Timeout = 60
	conf.TempPath = filepath.ToSlash(os.TempDir())
	conf.WebRoot = os.Getenv("WEB_ROOT")
	conf.Worker = 2
	conf.WebKitBin = os.Getenv("WEBKIT_BIN")
	conf.WebKitArgs = strings.Split(os.Getenv("WEBKIT_ARGS"), " ")

	http := NewHTTP(conf)

	http.Start()
}
