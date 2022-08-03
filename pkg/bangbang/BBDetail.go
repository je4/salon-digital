package bangbang

import (
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type BBDetail struct {
	*BangBang
	rootDir string
}

func (bb *BBDetail) SetRoutes(pathPrefix string, router *mux.Router) error {
	pathPrefix = strings.Trim(pathPrefix, "/")
	ps := strings.Split(pathPrefix, "/")
	bb.rootDir = ""
	for i := 0; i < len(ps); i++ {
		bb.rootDir += "../"
	}
	router.Handle("/{signature}", bb)
	return nil
}

func (bb *BBDetail) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bb.DetailHandler(w, r)
}
