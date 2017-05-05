package server

import (
	"errors"
	"github.com/tomogoma/go-commons/server/helper"
	"time"
	"os"
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

type Logger interface {
	Info(interface{}, ...interface{})
	Warn(interface{}, ...interface{}) error
	Error(interface{}, ...interface{}) error
}

type TokenValidator interface {
	Validate(token string, claims jwt.Claims) (*jwt.Token, error)
	IsAuthError(error) bool
}

type Config interface {
	ImagesDir() string
	ID() string
}

type Server struct {
	token   TokenValidator
	log     Logger
	tIDCh   chan int
	id      string
	imgsDir string
	model   Model
}

const (
	SomethingWickedError = "Something wicked happened"
)

var timeFormat = time.RFC3339

func New(c Config, tv TokenValidator, m Model, lg Logger) (*Server, error) {
	if tv == nil {
		return nil, errors.New("TokenValidator was nil")
	}
	if m == nil {
		return nil, errors.New("Model was nil")
	}
	if lg == nil {
		return nil, errors.New("Logger was nil")
	}
	if err := validateConfig(c); err != nil {
		return nil, err
	}
	tIDCh := make(chan int)
	go helper.TransactionSerializer(tIDCh)
	return &Server{id: c.ID(), imgsDir: c.ImagesDir(), model: m, token: tv, log: lg, tIDCh: tIDCh}, nil
}

func validateConfig(c Config) error {
	if c == nil {
		return errors.New("Config was nil")
	}
	if c.ID() == "" {
		return errors.New("ID was empty")
	}
	if _, err := os.Stat(c.ImagesDir()); err != nil {
		return fmt.Errorf("Unable to access images dir: %v", err)
	}
	return nil
}
