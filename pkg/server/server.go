package server

import (
	"context"
	"crypto/tls"
	"github.com/goph/emperror"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	dcert "github.com/je4/utils/v2/pkg/cert"
	"github.com/op/go-logging"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"time"
)

type SubServer interface {
	SetRoutes(pathPrefix string, route *mux.Router) error
}

type Server struct {
	service        string
	host, port     string
	name, password string
	srv            *http.Server
	linkTokenExp   time.Duration
	jwtKey         string
	jwtAlg         []string
	log            *logging.Logger
	urlExt         *url.URL
	accessLog      io.Writer
	subServer      map[string]SubServer
	dataFS         fs.FS
}

func NewServer(service, addr string, urlExt *url.URL, name, password string, dataFS fs.FS, log *logging.Logger, accessLog io.Writer) (*Server, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, emperror.Wrapf(err, "cannot split address %s", addr)
	}
	/*
		extUrl, err := url.Parse(addrExt)
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot parse external address %s", addrExt)
		}
	*/

	srv := &Server{
		service:   service,
		host:      host,
		port:      port,
		urlExt:    urlExt,
		dataFS:    dataFS,
		name:      name,
		password:  password,
		log:       log,
		accessLog: accessLog,
		subServer: map[string]SubServer{},
	}

	return srv, nil
}

func (s *Server) AddSubServer(path string, subServer SubServer) {
	s.subServer[path] = subServer
}

/*
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
		lab := salon.NewLabyrinth(7, signatures)

		if err := s.templates["grid.gohtml"].Execute(w, struct{ Lab *salon.Labyrinth }{Lab: lab}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%v", err)))
			return
		}
	}
}

*/
func (s *Server) ListenAndServe(cert, key string) (err error) {
	router := mux.NewRouter()
	/*
		injectFS, err := fs.Sub(web.InjectFS, "inject")
		if err != nil {
			return emperror.Wrap(err, "cannot get subtree of embedded inject")
		}
		httpInjectServer := http.FileServer(http.FS(injectFS))

		router.PathPrefix("/inject").Handler(
			http.StripPrefix("/inject", httpInjectServer),
		).Methods("GET")

		router.PathPrefix("/").Handler(
			http.StripPrefix("/", http.HandlerFunc(s.RegexpHandler)),
		).Methods("GET")
	*/
	router.PathPrefix("/data").Handler(http.StripPrefix("/data", http.FileServer(http.FS(s.dataFS)))).Methods("GET").Name("data server")

	for path, subServer := range s.subServer {
		subRouter := router.PathPrefix(path).Subrouter()
		subServer.SetRoutes(path, subRouter)
	}
	var tss = &TestSubServer{}
	tss.SetRoutes("/test", router.PathPrefix("/test").Subrouter())

	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathRegexp, err := route.GetPathRegexp()
		if err != nil {
			return emperror.Wrapf(err, "cannot get path regexp of route %s", route.GetName())
		}
		s.log.Infof("Route %s: %s", route.GetName(), pathRegexp)
		return nil
	})
	loggedRouter := handlers.CombinedLoggingHandler(s.accessLog, handlers.ProxyHeaders(router))
	addr := net.JoinHostPort(s.host, s.port)
	s.srv = &http.Server{
		Handler: loggedRouter,
		Addr:    addr,
	}

	if cert == "auto" || key == "auto" {
		s.log.Info("generating new certificate")
		cert, err := dcert.DefaultCertificate()
		if err != nil {
			return emperror.Wrap(err, "cannot generate default certificate")
		}
		s.srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{*cert}}
		s.log.Infof("starting salon digital at %v - https://%s:%v/", s.urlExt.String(), s.host, s.port)
		return s.srv.ListenAndServeTLS("", "")
	} else if cert != "" && key != "" {
		s.log.Infof("starting salon digital at %v - https://%s:%v/", s.urlExt.String(), s.host, s.port)
		return s.srv.ListenAndServeTLS(cert, key)
	} else {
		s.log.Infof("starting salon digital at %v - http://%s:%v/", s.urlExt.String(), s.host, s.port)
		return s.srv.ListenAndServe()
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
