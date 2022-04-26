package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type TestSubServer struct{}

func (tss *TestSubServer) SetRoutes(pathPrefix string, router *mux.Router) error {
	router.PathPrefix("/static").Handler(http.StripPrefix(pathPrefix+"/static", tss)).Methods("GET").Name("static test")
	//router.PathPrefix("/static").HandlerFunc(tss.MainHandler).Name("static test prefix")
	return nil
}

func (tss *TestSubServer) MainHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(r.URL.String())))
}

func (tss *TestSubServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tss.MainHandler(w, r)
}
