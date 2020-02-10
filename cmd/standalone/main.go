package main

import (
	"flag"
	"github.com/tomogoma/imagems/pkg/bootstrap"
	"github.com/tomogoma/imagems/pkg/config"
	"github.com/tomogoma/imagems/pkg/logging/logrus"
	http2 "net/http"
)

var confFilePath = flag.String(
	"conf",
	config.DefaultConfPath(),
	"path to config file",
)

func main() {
	log := &logrus.Wrapper{}
	flag.Parse()
	conf, err := config.ReadFile(*confFilePath)
	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
		return
	}
	handler, err := bootstrap.Bootstrap(log, *conf)
	err = http2.ListenAndServe(conf.Service.Address, handler)
	log.Fatalf("Quit with error: %v", err)
}
