// main
package main

import (
	"flag"
	"fmt"
	"runtime"

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

	service := lib.NewHTTP(conf)
	service.Start()
}
