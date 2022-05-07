package bangbang

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

type BBZoom struct {
	*BangBang
	rootDir string
}

func (bb *BBZoom) SetRoutes(pathPrefix string, router *mux.Router) error {
	pathPrefix = strings.Trim(pathPrefix, "/")
	ps := strings.Split(pathPrefix, "/")
	bb.rootDir = ""
	for i := 0; i < len(ps); i++ {
		bb.rootDir += "../"
	}
	router.Handle("/", bb)
	router.HandleFunc("/signature/{PosX}x{PosY}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pxs := vars["PosX"]
		pys := vars["PosY"]

		px, err := strconv.Atoi(pxs)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s is not a number: %v", pxs, err), http.StatusBadRequest)
			return
		}
		py, err := strconv.Atoi(pys)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s is not a number: %v", pys, err), http.StatusBadRequest)
			return
		}
		signature := bb.GetSignatureP(px, py)
		jsonEnc := json.NewEncoder(w)
		if err := jsonEnc.Encode(signature); err != nil {
			http.Error(w, fmt.Sprintf("cannot encode %s: %v", signature, err), http.StatusInternalServerError)
			return
		}
	})
	return nil
}

func (bb *BBZoom) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bb.ZoomHandler(w, r)
}
