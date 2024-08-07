package lib

import (
	"github.com/op/go-logging"
	"os"
)

var Logger = logging.MustGetLogger("html2pdf")

func init() {
	format := logging.MustStringFormatter(
		`HTML2PDF %{color} %{shortfunc} %{level:.4s} %{shortfile}
%{id:03x}%{color:reset} %{message}`,
	)
	logging.SetFormatter(format)

	levelStr := os.Getenv("LOG_LEVEL")
	if len(levelStr) == 0 {
		levelStr = "INFO"
	}
	level, err := logging.LogLevel(levelStr)
	if err != nil {
		level = logging.INFO
	}
	logging.SetLevel(level, "html2pdf")
}
