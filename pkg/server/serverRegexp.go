package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
)

var hostRegexp = regexp.MustCompile(`(?i)http://salon-digital\.zkm\.de`)
var headRegexp = regexp.MustCompile(`(?i)<HEAD>`)

func (s *Server) RegexpHandler(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
	}
	if !strings.HasSuffix(upath, ".html") && upath != "/" {
		s.httpStaticServer.ServeHTTP(w, r)
		return
	}
	if upath == "/" {
		upath = "/index.html"
	}
	upath = strings.TrimLeft(upath, "/")
	data, err := fs.ReadFile(s.staticFS, upath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("cannot read file %s", upath)))
	}
	w.Header().Set("Content-type", "text/html")
	data = hostRegexp.ReplaceAll(data, []byte(s.AddrExt))
	data = headRegexp.ReplaceAll(data, []byte(fmt.Sprintf("<HEAD>\n"+
		"<!-- CSS injections -->\n"+
		"<link rel=\"stylesheet\" type=\"text/css\" href=\"%s/inject/css/netscape.css\">\n"+
		"<!-- END CSS injection -->\n", s.AddrExt)))
	w.Write(data)
	return

}
