package server

import (
	"errors"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/go-commons/server/helper"
)

type Logger interface {
	Info(interface{}, ...interface{})
	Warn(interface{}, ...interface{}) error
	Error(interface{}, ...interface{}) error
}

type TokenValidator interface {
	Validate(token string) (*token.Token, error)
	IsClientError(error) bool
}

type Model interface {
}

type Server struct {
	token TokenValidator
	log   Logger
	tIDCh chan int
	id    string
}

const (
	SomethingWickedError = "Something wicked happened"
)

var ErrorNilTokenValidator = errors.New("TokenValidator was nil")
var ErrorNilLogger = errors.New("Logger was nil")
var ErrorEmptyID = errors.New("ID was empty")

func New(ID string, tv TokenValidator, lg Logger) (*Server, error) {
	if tv == nil {
		return nil, ErrorNilTokenValidator
	}
	if lg == nil {
		return nil, ErrorNilLogger
	}
	if ID == "" {
		return nil, ErrorEmptyID
	}
	tIDCh := make(chan int)
	go helper.TransactionSerializer(tIDCh)
	return &Server{id: ID, token: tv, log:lg, tIDCh: tIDCh}, nil
}
