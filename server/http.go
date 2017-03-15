package server

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) HandleRoute(r *mux.Router) {
	r.HandleFunc("/status", statusHandler)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		StatusCode int32
		Status     string
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"statusCode":200,"status":"ay-okay"}`))
}
