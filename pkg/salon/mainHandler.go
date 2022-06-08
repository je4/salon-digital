package salon

import (
	"fmt"
	"net/http"
	"strings"
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
	lab := NewLabyrinth(7, s.works)

	tpl, ok := s.gridTemplate["grid.gohtml"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("template grid.gohtml not found"))
		return
	}
	if err := tpl.Execute(w, struct {
		BaseAddr   string
		Lab        *Labyrinth
		Responsive bool
	}{
		BaseAddr:   strings.TrimRight(s.addrExt, "/") + "/" + strings.Trim(s.pathPrefix, "/") + "/",
		Lab:        lab,
		Responsive: s.responsive,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
}
