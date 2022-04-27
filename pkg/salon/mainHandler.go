package salon

import (
	"fmt"
	"net/http"
)

func (s *Salon) MainHandler(w http.ResponseWriter, r *http.Request) {
	if s.templateDev {
		if err := s.initTemplates(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%v", err)))
			return
		}
	}
	var signatures = []string{}
	for key, _ := range s.works {
		signatures = append(signatures, key)
	}
	lab := NewLabyrinth(7, signatures)

	tpl, ok := s.gridTemplate["grid.gohtml"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("template grid.gohtml not found"))
		return
	}
	if err := tpl.Execute(w, struct{ Lab *Labyrinth }{Lab: lab}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
}
