package salon

import (
	"github.com/Masterminds/sprig"
	"github.com/blevesearch/bleve/v2"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

type Salon struct {
	index            bleve.Index
	gridTemplate     *template.Template
	templateFS       fs.FS
	gridTemplateFile string
	log              *logging.Logger
	httpImageServer  http.Handler
}

func NewSalon(index bleve.Index, templatePath, gridTemplateFile string, imageFS fs.FS, log *logging.Logger) (*Salon, error) {
	var templateFS fs.FS
	if templatePath == "" {
		templateFS = TemplateFS
	} else {
		templateFS = os.DirFS(templatePath)
	}
	s := &Salon{
		index:            index,
		gridTemplate:     nil,
		templateFS:       templateFS,
		gridTemplateFile: gridTemplateFile,
		log:              log,
		httpImageServer:  http.FileServer(http.FS(imageFS)),
	}
	return s, s.initTemplates()
}

func (s *Salon) initTemplates() (err error) {
	funcMap := sprig.FuncMap()
	funcMap["iterate"] = func(count int) []int {
		var i int
		var Items []int
		for i = 0; i < count; i++ {
			Items = append(Items, i)
		}
		return Items
	}
	name := filepath.Base(s.gridTemplateFile)
	s.gridTemplate, err = template.New(name).Funcs(funcMap).ParseFS(s.templateFS, s.gridTemplateFile)
	if err != nil {
		return errors.Wrapf(err, "cannot parse template: %s", name)
	}
	return nil
}

func (s *Salon) SetRoutes(router *mux.Router) error {
	router.PathPrefix("/img").Handler(http.StripPrefix("/img", s.httpImageServer)).Methods("GET")
	router.PathPrefix("/").HandlerFunc(s.MainHandler).Methods("GET")
	router.HandleFunc("/move/{direction}", s.MoveHandler).Methods("POST")
	return nil
}
