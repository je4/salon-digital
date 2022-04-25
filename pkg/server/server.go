package server

import (
	"context"
	"crypto/tls"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	dcert "github.com/je4/utils/v2/pkg/cert"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type SubServer interface {
	SetRoutes(route *mux.Router) error
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
	AddrExt        string
	accessLog      io.Writer
	subServer      map[string]SubServer
}

func NewServer(service,
	addr, addrExt string,
	name, password string,
	log *logging.Logger,
	accessLog io.Writer) (*Server, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot split address %s", addr)
	}
	/*
		extUrl, err := url.Parse(addrExt)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse external address %s", addrExt)
		}
	*/

	srv := &Server{
		service:   service,
		host:      host,
		port:      port,
		AddrExt:   strings.TrimRight(addrExt, "/"),
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
			return errors.Wrap(err, "cannot get subtree of embedded inject")
		}
		httpInjectServer := http.FileServer(http.FS(injectFS))

		router.PathPrefix("/inject").Handler(
			http.StripPrefix("/inject", httpInjectServer),
		).Methods("GET")

		router.PathPrefix("/").Handler(
			http.StripPrefix("/", http.HandlerFunc(s.RegexpHandler)),
		).Methods("GET")
	*/
	for path, subServer := range s.subServer {
		subRouter := router.PathPrefix(path).Subrouter()
		subServer.SetRoutes(subRouter)
	}
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
			return errors.Wrap(err, "cannot generate default certificate")
		}
		s.srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{*cert}}
		s.log.Infof("starting salon digital at %v - https://%s:%v/", s.AddrExt, s.host, s.port)
		return s.srv.ListenAndServeTLS("", "")
	} else if cert != "" && key != "" {
		s.log.Infof("starting salon digital at %v - https://%s:%v/", s.AddrExt, s.host, s.port)
		return s.srv.ListenAndServeTLS(cert, key)
	} else {
		s.log.Infof("starting salon digital at %v - http://%s:%v/", s.AddrExt, s.host, s.port)
		return s.srv.ListenAndServe()
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
