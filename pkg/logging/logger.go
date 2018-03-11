package logging

import (
	"net/http"
)

type Logger interface {
	WithHTTPRequest(r *http.Request) Logger
	WithFields(map[string]interface{}) Logger
	WithField(string, interface{}) Logger
	Infof(fmt string, args ...interface{})
	Warnf(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}
