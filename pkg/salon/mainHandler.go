package salon

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"net/http"
)

func (s *Salon) MainHandler(w http.ResponseWriter, r *http.Request) {
	bQuery := bleve.NewMatchAllQuery()
	bSearch := bleve.NewSearchRequest(bQuery)
	searchResult, err := s.index.Search(bSearch)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
	var signatures = []string{}
	for _, val := range searchResult.Hits {
		signatures = append(signatures, val.ID)
	}
	lab := NewLabyrinth(7, signatures)

	if err := s.gridTemplate.Execute(w, struct{ Lab *Labyrinth }{Lab: lab}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
}
