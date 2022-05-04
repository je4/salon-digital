package bangbang

type BBServer struct{}

/*
func (tss *BBServer) SetRoutes(pathPrefix string, router *mux.Router) error {
	router.Handle("/{signature}", tss.MainHandler)
	//router.PathPrefix("/static").HandlerFunc(tss.MainHandler).Name("static test prefix")
	return nil
}

func (tss *BBServer) MainHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(r.URL.String())))
}

func (tss *BBServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tss.MainHandler(w, r)
}

*/
