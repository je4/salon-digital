package bangbang

import (
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type BBDocs struct {
	*BangBang
	rootDir string
}

func (bb *BBDocs) SetRoutes(pathPrefix string, router *mux.Router) error {
	pathPrefix = strings.Trim(pathPrefix, "/")
	ps := strings.Split(pathPrefix, "/")
	bb.rootDir = ""
	for i := 0; i < len(ps); i++ {
		bb.rootDir += "../"
	}
	router.Handle("/{signature}", bb)
	return nil
}

func (bb *BBDocs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bb.DocumentHandler(w, r)
}
