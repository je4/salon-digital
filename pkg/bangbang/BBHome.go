package bangbang

import (
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type BBHome struct {
	*BangBang
	rootDir string
}

func (bb *BBHome) SetRoutes(pathPrefix string, router *mux.Router) error {
	pathPrefix = strings.Trim(pathPrefix, "/")
	ps := strings.Split(pathPrefix, "/")
	bb.rootDir = ""
	for i := 0; i < len(ps); i++ {
		bb.rootDir += "../"
	}
	router.Handle("/", bb)
	return nil
}

func (bb *BBHome) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bb.HomeHandler(w, r)
}
