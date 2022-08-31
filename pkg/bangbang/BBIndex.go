package bangbang

import (
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type BBIndex struct {
	*BangBang
	rootDir string
}

func (bb *BBIndex) SetRoutes(pathPrefix string, router *mux.Router) error {
	pathPrefix = strings.Trim(pathPrefix, "/")
	ps := strings.Split(pathPrefix, "/")
	bb.rootDir = ""
	for i := 0; i < len(ps); i++ {
		bb.rootDir += "../"
	}
	router.Handle("/{type}", bb)
	return nil
}

func (bb *BBIndex) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bb.IndexHandler(w, r)
}
