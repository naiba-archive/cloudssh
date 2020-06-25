package main

import (
	"flag"

	"github.com/naiba/cloudssh/cmd/server/router"
)

func main() {
	var conf string
	flag.StringVar(&conf, "conf", "config.json", "config file path")
	flag.Parse()
	router.Serve(conf, 80)
}
