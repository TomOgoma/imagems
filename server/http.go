package server

import (
	"net/http"
	"github.com/gorilla/mux"
)

func (s *Server) NewHttpHandler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/status", statusHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(s.imgsDir))).Methods("GET")
	return r
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		StatusCode int32
		Status     string
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"statusCode":200,"status":"ay-okay"}`))
}
