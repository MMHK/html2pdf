// main
package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	"html2pdf/lib"
)

func main() {
	conf_path := flag.String("c", "config.json", "config json file")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	err, conf := lib.NewConfig(*conf_path)
	if err != nil {
		fmt.Println(err)
		return
	}
	conf = conf.LoadWithENV()

	cleaner := lib.NewCleaner(time.Duration(conf.Cleaner.CleanupPeriod)*time.Second,
		time.Duration(conf.Cleaner.FileAgeLimit)*time.Second)

	cleaner.Start();
	defer cleaner.Stop();

	service := lib.NewHTTP(conf)
	service.Start()
}
