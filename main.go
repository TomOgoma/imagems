package main

import (
	"flag"
	"time"
	"runtime"
	"fmt"
	"github.com/limetext/log4go"
	"github.com/micro/go-micro"
	"github.com/tomogoma/imagems/server"
	"github.com/tomogoma/go-commons/auth/token"
	confhelper "github.com/tomogoma/go-commons/config"
	"github.com/tomogoma/imagems/config"
	"github.com/micro/go-web"
	"path"
	"github.com/tomogoma/imagems/model"
	"github.com/tomogoma/imagems/db"
	"os"
	"io/ioutil"
	"strings"
	"github.com/tomogoma/imagems/server/proto"
)

type Logger interface {
	Fine(interface{}, ...interface{})
	Info(interface{}, ...interface{})
	Warn(interface{}, ...interface{}) error
	Error(interface{}, ...interface{}) error
}

type FileWriter func(fileName string, data []byte, perm os.FileMode) error

func (f FileWriter) WriteFile(fName string, dt []byte, perm os.FileMode) error {
	return f(fName, dt, perm)
}

type ServiceConfig config.Service

func (sc ServiceConfig) ImagesDir() string {
	return path.Join(sc.DataDir, imgsDirName)
}

func (sc ServiceConfig) ImgURLRoot() string {
	if strings.HasSuffix(sc.ImgURL, "/") {
		return sc.ImgURL + name
	}
	return sc.ImgURL + "/" + name
}

func (sc ServiceConfig) ID() string {
	return apiID
}

const (
	name = "imagems"
	apiID = "go.micro.api." + name
	webID = "go.micro.web." + name
	version = "0.1.0"
	confCommand = "conf"
	defaultConfFile = "/etc/" + name + "/" + name + ".conf.yaml"
	imgsDirName = "images"
)

var confFilePath = flag.String(confCommand, defaultConfFile, "path to config file")

func main() {
	flag.Parse();
	defer func() {
		runtime.Gosched()
		time.Sleep(50 * time.Millisecond)
	}()
	conf := config.Config{}
	log := log4go.NewDefaultLogger(log4go.FINEST)
	err := confhelper.ReadYamlConfig(*confFilePath, &conf)
	if err != nil {
		log.Critical("Error reading config file: %s", err)
		return
	}
	err = bootstrap(log, conf)
	log.Critical("Quit with error: %v", err)
}

// bootstrap collects all the dependencies necessary to start the server,
// injects said dependencies, and proceeds to register it as a micro grpc handler.
func bootstrap(log Logger, conf config.Config) error {
	tv, err := token.NewGenerator(conf.Auth)
	if err != nil {
		return fmt.Errorf("Error instantiating token validator: %s", err)
	}
	d, err := db.New(conf.Database)
	if err != nil {
		return fmt.Errorf("Error instantiating database: %v", err)
	}
	m, err := model.New(ServiceConfig(conf.Service), d, FileWriter(ioutil.WriteFile))
	if err != nil {
		return fmt.Errorf("Error instantiating model: %v", err)
	}
	srv, err := server.New(ServiceConfig(conf.Service), tv, m, log);
	if err != nil {
		return fmt.Errorf("Error instantiating rpc server: %s", err)
	}
	rpc := micro.NewService(
		micro.Name(apiID),
		micro.Version(version),
		micro.RegisterInterval(conf.Service.RegisterInterval),
	)
	image.RegisterImageHandler(rpc.Server(), srv)
	wb := web.NewService(
		web.Name(webID),
		web.Version(version),
		web.RegisterInterval(conf.Service.RegisterInterval),
	)
	wb.Handle("/", srv.NewHttpHandler())
	errCH := make(chan error)
	go func() {
		if err := rpc.Run(); err != nil {
			errCH <- fmt.Errorf("Error serving rpc: %s", err)
		}
		errCH <- nil
	}()
	go func() {
		if err := wb.Run(); err != nil {
			errCH <- fmt.Errorf("Error serving web: %s", err)
		}
		errCH <- nil
	}()
	return <-errCH
}
