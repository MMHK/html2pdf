// main
package main

import (
	"HTML2PDF/lib"
	"flag"
	"fmt"
	"runtime"
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

	service := lib.NewHTTP(conf)
	service.Start()
}
