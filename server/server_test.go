package server_test

import (
	"testing"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/imagems/server"
	"github.com/limetext/log4go"
	"time"
	"os"
)

type ConfigMock struct {
	ExpImDir string
	ExpSrvID string
}

func (c *ConfigMock) ImagesDir() string {
	return c.ExpImDir
}
func (c *ConfigMock) ID() string {
	return c.ExpSrvID
}

type TokenValidatorMock struct {
	ExpToken   *token.Token
	ExpValErr  error
	ExpIsClErr bool
}

func (t *TokenValidatorMock) Validate(token string) (*token.Token, error) {
	return t.ExpToken, t.ExpValErr
}

func (t *TokenValidatorMock) IsClientError(error) bool {
	return t.ExpIsClErr
}

type ModelMock struct {
	ExpImURL        string
	ExpNewImErr     error
	ExpUnauthorized bool
	ExpIsClErr      bool
}

func (m *ModelMock) NewImage(*token.Token, int64, []byte) (string, string, error) {
	return srvTm, m.ExpImURL, m.ExpNewImErr
}
func (m *ModelMock) IsUnauthorizedError(error) bool {
	return m.ExpUnauthorized
}
func (m *ModelMock) IsClientError(error) bool {
	return m.ExpIsClErr
}

const srvID = "go.micro.api.imagems_tests"
const imgDir = "test/images"

var logger = log4go.Logger{}
var srvTm = time.Now().Format(time.RFC3339)
var validConfig = &ConfigMock{ExpSrvID: srvID, ExpImDir: imgDir}

func TestNew(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	s, err := server.New(validConfig, &TokenValidatorMock{}, &ModelMock{}, logger)
	if err != nil {
		t.Fatalf("server.New(): %v", err)
	}
	if s == nil {
		t.Fatal("Got a nil Server")
	}
}

func TestNew_nilConfig(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	_, err := server.New(nil, &TokenValidatorMock{}, &ModelMock{}, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_emptyID(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	_, err := server.New(&ConfigMock{ExpImDir: imgDir}, &TokenValidatorMock{}, &ModelMock{}, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_NonExistImgDir(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	_, err := server.New(&ConfigMock{ExpSrvID: srvID, ExpImDir: "test/some/non-exist/img/dir"},
		&TokenValidatorMock{}, &ModelMock{}, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

// this test will fail if run as root
func TestNew_uncreateableDir(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	_, err := server.New(&ConfigMock{ExpImDir:"/tmp/imagems_test"}, &TokenValidatorMock{}, &ModelMock{}, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_nilTokenValidator(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	_, err := server.New(validConfig, nil, &ModelMock{}, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_nilModel(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	_, err := server.New(validConfig, &TokenValidatorMock{}, nil, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_nilLogger(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	_, err := server.New(validConfig, &TokenValidatorMock{}, &ModelMock{}, nil)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func setUp(t *testing.T) {
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		t.Fatalf("Error setting up (creating test images dir): %v", err)
	}
}

func tearDown(t *testing.T) {
	err := os.RemoveAll("test")
	if err != nil {
		t.Logf("Error tearing down: %v", err)
	}
}
