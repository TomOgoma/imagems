package config

import (
	"github.com/tomogoma/go-commons/auth/token"
	"time"
	"github.com/tomogoma/go-commons/database/cockroach"
)

type Service struct {
	RegisterInterval time.Duration `yaml:"registerInterval,omitempty"`
	DataDir          string `yaml:"dataDir,omitempty"`
	ImgURL           string `yaml:"imgURL,omitempty"`
}

type Config struct {
	Auth     token.ConfigStub `yaml:"auth,omitempty"`
	Service  Service `yaml:"service,omitempty"`
	Database cockroach.DSN `yaml:"database"`
}
