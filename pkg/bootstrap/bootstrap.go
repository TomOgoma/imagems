package bootstrap

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/imagems/pkg/api"
	"github.com/tomogoma/imagems/pkg/config"
	"github.com/tomogoma/imagems/pkg/handler/http"
	"github.com/tomogoma/imagems/pkg/jwt"
	"github.com/tomogoma/imagems/pkg/logging"
	"github.com/tomogoma/imagems/pkg/model"
	"github.com/tomogoma/imagems/pkg/roach"
	jwt2 "github.com/tomogoma/jwt"
	netHttp "net/http"
)

// bootstrap collects all the dependencies necessary to start the server,
// injects said dependencies, and proceeds to register it as a micro grpc handler.
func Bootstrap(log logging.Logger, conf config.Config) (netHttp.Handler, error) {

	jwtKey, err := ioutil.ReadFile(conf.Auth.TokenKeyFile)
	if err != nil {
		return nil, errors.Newf("read auth token key file: %v", err)
	}
	jwter, err := jwt2.NewHandler(jwtKey)
	if err != nil {
		return nil, errors.Newf("new jwt handler: %v", err)
	}
	tknVal, err := jwt.NewValidator(jwter)
	if err != nil {
		return nil, errors.Newf("new jwt validator: %v", err)
	}

	d := roach.New(
		roach.WithDSN(conf.Database.FormatDSN()),
		roach.WithDBName(conf.Database.DBName),
	)
	if err := d.InitDBIfNot(); err != nil {
		log.Warnf("Unable to initialize database connection: %v", err)
	}

	m, err := model.New(conf.Service, tknVal, d, FileWriter(ioutil.WriteFile))
	if err != nil {
		return nil, fmt.Errorf("new model: %v", err)
	}

	genAPIKey, err := ioutil.ReadFile(conf.Auth.GenAPIKeyFile)
	if err != nil {
		log.Warnf("No general API key found: %v", err)
		genAPIKey = []byte{}
	}
	g, err := api.NewGuard(d, api.WithMasterKey(string(genAPIKey)))

	handler, err := http.NewHandler(conf.Service, m, g, log)
	if err != nil {
		return nil, fmt.Errorf("new HTTP handler: %s", err)
	}
	return handler, nil
}

type FileWriter func(fileName string, data []byte, perm os.FileMode) error

func (f FileWriter) WriteFile(fName string, dt []byte, perm os.FileMode) error {
	return f(fName, dt, perm)
}
