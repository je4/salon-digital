package bangbang

import (
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type BBGrid struct {
	*BangBang
	rootDir string
}

func (bb *BBGrid) SetRoutes(pathPrefix string, router *mux.Router) error {
	pathPrefix = strings.Trim(pathPrefix, "/")
	ps := strings.Split(pathPrefix, "/")
	bb.rootDir = ""
	for i := 0; i < len(ps); i++ {
		bb.rootDir += "../"
	}
	router.Handle("/", bb)
	return nil
}

func (bb *BBGrid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bb.GridHandler(w, r)
}
