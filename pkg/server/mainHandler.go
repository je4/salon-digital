package server

import (
	"fmt"
	"net/http"
)

func (s *Server) MainHandler(w http.ResponseWriter, r *http.Request) {
	if s.templateDev {
		if err := s.InitTemplates(); err != nil {
			s.log.Errorf("error initializing templates: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%v", err)))
			return
		}
	}
	url := r.URL
	if url.Path == "/" {
		lab := NewLabyrinth(7)

		if err := s.templates["grid.gohtml"].Execute(w, struct{ Lab *Labyrinth }{Lab: lab}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%v", err)))
			return
		}
	}
}
