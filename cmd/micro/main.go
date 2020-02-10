package main

import (
	"flag"
	"github.com/micro/go-web"
	"github.com/tomogoma/imagems/pkg/bootstrap"
	"github.com/tomogoma/imagems/pkg/config"
	"github.com/tomogoma/imagems/pkg/logging/logrus"
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
	err = web.NewService(
		web.Handler(handler),
		web.Name(config.CanonicalWebName()),
		web.Version(conf.Service.LoadBalanceVersion),
		web.RegisterInterval(conf.Service.RegisterInterval),
	).Run()
	log.Fatalf("Quit with error: %v", err)
}
