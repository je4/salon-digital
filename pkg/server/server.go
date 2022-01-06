package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/je4/salon-digital/v2/web"
	dcert "github.com/je4/utils/v2/pkg/cert"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"html/template"
	"io"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	service          string
	host, port       string
	name, password   string
	srv              *http.Server
	linkTokenExp     time.Duration
	jwtKey           string
	jwtAlg           []string
	log              *logging.Logger
	AddrExt          string
	accessLog        io.Writer
	templates        map[string]*template.Template
	httpStaticServer http.Handler
	staticFS         fs.FS
}

func NewServer(service, addr, addrExt, name, password string, log *logging.Logger, accessLog io.Writer) (*Server, error) {
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
		templates: map[string]*template.Template{},
	}
	srv.staticFS, err = fs.Sub(web.StaticFS, "static")
	if err != nil {
		return nil, errors.Wrap(err, "cannot get subtree of embedded static")
	}
	srv.httpStaticServer = http.FileServer(http.FS(srv.staticFS))

	return srv, srv.InitTemplates()
}

func (s *Server) InitTemplates() error {
	entries, err := web.TemplateFS.ReadDir("template")
	if err != nil {
		return errors.Wrapf(err, "cannot read template folder %s", "template")
	}
	for _, entry := range entries {
		name := entry.Name()
		tpl, err := template.ParseFS(web.TemplateFS, "template/"+name)
		if err != nil {
			return errors.New(fmt.Sprintf("cannot parse template: %s", name))
		}
		s.templates[name] = tpl
	}
	return nil
}

func (s *Server) ListenAndServe(cert, key string) (err error) {
	router := mux.NewRouter()

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
