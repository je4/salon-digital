package server

import (
	"fmt"
	"net/http"
)

func (s *Server) MainHandler(w http.ResponseWriter, r *http.Request) {
	if s.templateDev {
		s.InitTemplates()
	}
	url := r.URL
	if url.Path == "/" {
		if err := s.templates["grid.gohtml"].Execute(w, nil); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%v", err)))
			return
		}
	}
}
