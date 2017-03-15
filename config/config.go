package config

import (
	"github.com/tomogoma/go-commons/auth/token"
	"time"
)

type Service struct {
	RegisterInterval time.Duration `yaml:"registerInterval,omitempty"`
	DataDir          string `yaml:"dataDir,omitempty"`
}

type Config struct {
	Auth    token.DefaultConfig `yaml:"auth,omitempty"`
	Service Service `yaml:"service,omitempty"`
}
