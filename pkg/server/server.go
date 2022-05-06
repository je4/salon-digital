package server

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
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

var expectedUsernameHash = sha256.Sum256([]byte("bangbang"))
var expectedPasswordHash = sha256.Sum256([]byte("2022"))

//func basicAuth(next http.HandlerFunc) http.HandlerFunc {
func basicAuth(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// https://www.alexedwards.net/blog/basic-authentication-in-go
		// License: MIT
		// Extract the username and password from the request
		// Authorization header. If no Authentication header is present
		// or the header value is invalid, then the 'ok' return value
		// will be false.
		username, password, ok := r.BasicAuth()
		if ok {
			// Calculate SHA-256 hashes for the provided and expected
			// usernames and passwords.
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))

			// Use the subtle.ConstantTimeCompare() function to check if
			// the provided username and password hashes equal the
			// expected username and password hashes. ConstantTimeCompare
			// will return 1 if the values are equal, or 0 otherwise.
			// Importantly, we should to do the work to evaluate both the
			// username and password before checking the return values to
			// avoid leaking information.
			usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
			passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

			// If the username and password are correct, then call
			// the next handler in the chain. Make sure to return
			// afterwards, so that none of the code below is run.
			if usernameMatch && passwordMatch {
				h.ServeHTTP(w, r)
				return
			}
		}

		// If the Authentication header is not present, is invalid, or the
		// username or password is wrong, then set a WWW-Authenticate
		// header to inform the client that we expect them to use basic
		// authentication and send a 401 Unauthorized response.
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
	return http.HandlerFunc(fn)
}

func (s *Server) AddSubServer(path string, subServer SubServer) {
	s.subServer[path] = subServer
}

func (s *Server) ListenAndServe(cert, key string) (err error) {
	router := mux.NewRouter()

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
	loggedRouter := handlers.CombinedLoggingHandler(s.accessLog, handlers.ProxyHeaders(basicAuth(router)))
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
