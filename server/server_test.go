package server_test

import (
	"testing"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/imagems/server"
	"github.com/limetext/log4go"
)

type TokenValidatorMock struct {
	ExpToken *token.Token
	ExpErr   error
	ExpClErr bool
}

func (t *TokenValidatorMock) Validate(token string) (*token.Token, error) {
	return t.ExpToken, t.ExpErr
}

func (t *TokenValidatorMock) IsClientError(error) bool {
	return t.ExpClErr
}

var srvID = "test_server"
var logger = log4go.Logger{}

func TestNew(t *testing.T) {
	s, err := server.New(srvID, &TokenValidatorMock{}, logger)
	if err != nil {
		t.Fatalf("server.New(): %v", err)
	}
	if s == nil {
		t.Fatal("Got a nil Server")
	}
}

func TestNew_emptyID(t *testing.T) {
	_, err := server.New("", &TokenValidatorMock{}, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_nilTokenValidator(t *testing.T) {
	_, err := server.New(srvID, nil, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_nilLogger(t *testing.T) {
	_, err := server.New(srvID, &TokenValidatorMock{}, nil)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}
