package config

import (
	"time"
	"github.com/tomogoma/crdb"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
)

type Auth struct {
	TokenKeyFile string `json:"tokenKeyFile" yaml:"tokenKeyFile"`
}

type Service struct {
	RegisterInterval time.Duration `yaml:"registerInterval,omitempty"`
	DataDir          string        `yaml:"dataDir,omitempty"`
	ImgURL           string        `yaml:"imgURL,omitempty"`
}

type Config struct {
	Auth     Auth        `yaml:"auth,omitempty"`
	Service  Service     `yaml:"service,omitempty"`
	Database crdb.Config `yaml:"database"`
}

func ReadFile(fName string) (*Config, error) {
	confD, err := ioutil.ReadFile(fName)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	if err := yaml.Unmarshal(confD, conf); err != nil {
		return nil, fmt.Errorf("unmarshal yaml file (%s): %v",
			fName, err)
	}
	return conf, nil
}
