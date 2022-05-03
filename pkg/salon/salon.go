package salon

import (
	"github.com/Masterminds/sprig"
	"github.com/goph/emperror"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
)

type Salon struct {
	works        map[string]*Work
	gridTemplate map[string]*template.Template
	templateFS   fs.FS
	staticFS     fs.FS
	imageFS      fs.FS
	log          *logging.Logger
	templateDev  bool
}

func NewSalon(works map[string]*Work, staticFS, templateFS fs.FS, templateDev bool, imageFS fs.FS, log *logging.Logger) (*Salon, error) {
	s := &Salon{
		works:        works,
		gridTemplate: map[string]*template.Template{},
		templateFS:   templateFS,
		templateDev:  templateDev,
		staticFS:     staticFS,
		imageFS:      imageFS,
		log:          log,
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
		return emperror.Wrap(err, "cannot find templates")
	}
	for _, templateFile := range templateFiles {
		name := filepath.Base(templateFile)
		s.gridTemplate[name], err = template.New(name).Funcs(funcMap).ParseFS(s.templateFS, templateFile)
		if err != nil {
			return emperror.Wrapf(err, "cannot parse template: %s", templateFile)
		}
	}
	return nil
}

func (s *Salon) SetRoutes(pathPrefix string, router *mux.Router) error {
	router.PathPrefix("/img").Handler(http.StripPrefix(pathPrefix+"/img", http.FileServer(http.FS(s.imageFS)))).Methods("GET").Name("image server")
	router.PathPrefix("/static").Handler(http.StripPrefix(pathPrefix+"/static", http.FileServer(http.FS(s.staticFS)))).Methods("GET").Name("static server")
	router.HandleFunc("/", s.MainHandler).Methods("GET").Name("main")
	router.HandleFunc("/move/{direction}", s.MoveHandler).Methods("POST").Name("move")
	return nil
}
