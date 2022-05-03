package salon

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Salon) MoveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	var lab = &Labyrinth{}
	if err := json.NewDecoder(r.Body).Decode(lab); err != nil {
		s.log.Errorf("error decoding body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		jw := json.NewEncoder(w)
		jw.Encode(struct{ error string }{error: fmt.Sprintf("error decoding body: %v", err)})
		return
	}
	lab.initData(s.works)
	direction := vars["direction"]
	switch direction {
	case "east":
		lab.Move(-1, 0)
	case "west":
		lab.Move(1, 0)
	case "south":
		lab.Move(0, -1)
	case "north":
		lab.Move(0, 1)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("invalid direction: %s", direction)))
		return
	}
	jenc := json.NewEncoder(w)
	//jenc.SetIndent("", "  ")
	if err := jenc.Encode(lab); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errStr := fmt.Sprintf("internal error encoding json: %v", err)
		w.Write([]byte(errStr))
		s.log.Error(errStr)
		return
	}
}
