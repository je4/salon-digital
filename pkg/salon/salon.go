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
	"path/filepath"
)

type Salon struct {
	index            bleve.Index
	gridTemplate     map[string]*template.Template
	templateFS       fs.FS
	staticFS         fs.FS
	log              *logging.Logger
	httpImageServer  http.Handler
	httpStaticServer http.Handler
	templageDev      bool
}

func NewSalon(index bleve.Index, staticFS, templateFS fs.FS, templateDev bool, imageFS fs.FS, log *logging.Logger) (*Salon, error) {
	s := &Salon{
		index:            index,
		gridTemplate:     map[string]*template.Template{},
		templateFS:       templateFS,
		templageDev:      templateDev,
		staticFS:         staticFS,
		log:              log,
		httpImageServer:  http.FileServer(http.FS(imageFS)),
		httpStaticServer: http.FileServer(http.FS(staticFS)),
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
	templateFiles, err := findAllFiles(s.templateFS, ".", ".gohtml")
	if err != nil {
		return errors.Wrap(err, "cannot find templates")
	}
	for _, templateFile := range templateFiles {
		name := filepath.Base(templateFile)
		s.gridTemplate[name], err = template.New(name).Funcs(funcMap).ParseFS(s.templateFS, templateFile)
		if err != nil {
			return errors.Wrapf(err, "cannot parse template: %s", templateFile)
		}
	}
	return nil
}

func (s *Salon) SetRoutes(pathPrefix string, router *mux.Router) error {
	router.PathPrefix("/img").Handler(http.StripPrefix(pathPrefix+"/img", s.httpImageServer)).Methods("GET").Name("image server")
	router.PathPrefix("/static").Handler(http.StripPrefix(pathPrefix+"/static", s.httpStaticServer)).Methods("GET").Name("static server")
	router.HandleFunc("/", s.MainHandler).Methods("GET").Name("main")
	router.HandleFunc("/move/{direction}", s.MoveHandler).Methods("POST").Name("move")
	return nil
}
