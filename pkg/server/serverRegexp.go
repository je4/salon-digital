package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
)

var hostSalonRegexp = regexp.MustCompile(`(?i)http://salon-digital\.zkm\.de`)
var hostWWW2Regexp = regexp.MustCompile(`(?i)http://www2\.zkm\.de`)
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
		http.Redirect(w, r, s.AddrExt+"/salon/index.html", 301)
		return
		//upath = "/salon/index.html"
	}
	upath = strings.TrimLeft(upath, "/")
	data, err := fs.ReadFile(s.staticFS, upath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("cannot read file %s", upath)))
	}
	w.Header().Set("Content-type", "text/html")
	data = hostSalonRegexp.ReplaceAll(data, []byte(s.AddrExt+"/salon"))
	data = hostWWW2Regexp.ReplaceAll(data, []byte(s.AddrExt+"/www2"))
	if !strings.HasSuffix(upath, "login.html") {
		data = headRegexp.ReplaceAll(data, []byte(fmt.Sprintf("<HEAD>\n"+
			"<!-- BEGIN CSS injections -->\n"+
			"<link rel=\"stylesheet\" type=\"text/css\" href=\"%s/inject/css/oldsalon.css\">\n"+
			"<script src=\"%s/inject/js/oldsalon.js\"></script>\n"+
			"<script>\nwindow.onload = function() {\n  initNetscape();\n};\n</script>"+
			"<!-- END CSS injection -->\n", s.AddrExt, s.AddrExt)))
	}
	w.Write(data)
	return

}
