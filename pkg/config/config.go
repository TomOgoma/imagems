package config

import (
	"time"
	"github.com/tomogoma/crdb"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"path"
	"strings"
)

type Auth struct {
	TokenKeyFile  string `json:"tokenKeyFile" yaml:"tokenKeyFile"`
	GenAPIKeyFile string `json:"genAPIKeyFile" yaml:"genAPIKeyFile"`
}

type Service struct {
	RegisterInterval   time.Duration `yaml:"registerInterval" json:"registerInterval"`
	DataDir            string        `yaml:"dataDir" json:"dataDir"`
	ImgURL             string        `yaml:"imgURL" json:"imgURL"`
	LoadBalanceVersion string        `yaml:"loadBalanceVersion" json:"loadBalanceVersion"`
}

func (sc Service) ImagesDir() string {
	return path.Join(sc.DataDir, imgsDirName)
}

func (sc Service) DefaultFolderName() string {
	return "general"
}

func (sc Service) ImgURLRoot() string {
	return strings.TrimSuffix(sc.ImgURL, "/") + WebRootURL()
}

type Config struct {
	Auth     Auth        `yaml:"auth" json:"auth"`
	Service  Service     `yaml:"service" json:"service"`
	Database crdb.Config `yaml:"database" json:"database"`
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
