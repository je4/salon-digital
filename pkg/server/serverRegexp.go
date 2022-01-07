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
var endBodyRegexp = regexp.MustCompile(`(?i)</BODY>`)
var startFont = regexp.MustCompile(`(?i)<font`)
var endFont = regexp.MustCompile(`(?i)</font`)

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
		"<!-- BEGIN CSS injections -->\n"+
		"<link rel=\"stylesheet\" type=\"text/css\" href=\"%s/inject/css/netscape.css\">\n"+
		"<script src=\"%s/inject/js/netscape.js\"></script>\n"+
		"<script>\nwindow.onload = function() {\n  initNetscape();\n};\n</script>"+
		"<!-- END CSS injection -->\n", s.AddrExt, s.AddrExt)))
	/*
		data = endBodyRegexp.ReplaceAll(data, []byte(fmt.Sprintf("<!-- BEGIN Javascript injections -->\n"+
			"<link rel=\"stylesheet\" type=\"text/css\" href=\"%s/inject/css/netscape.css\">\n"+
			"<!-- END Javascript injection -->\n"+
			"</BODY>", s.AddrExt)))
	*/
	//	data = startFont.ReplaceAll(data, []byte("<x-font"))
	//		data = endFont.ReplaceAll(data, []byte("</x-font"))

	w.Write(data)
	return

}
