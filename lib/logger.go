package lib

import (
	"log"
	"os"
)

var (
	InfoLogger = log.New(os.Stdout, "INFO|", log.LstdFlags|log.Lshortfile)
	ErrLogger  = log.New(os.Stderr, "ERROR|", log.LstdFlags|log.Lshortfile)
)
