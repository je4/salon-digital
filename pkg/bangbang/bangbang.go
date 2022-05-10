package bangbang

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/blevesearch/bleve/v2"
	"github.com/goph/emperror"
	"github.com/gorilla/mux"
	"github.com/je4/salon-digital/v2/pkg/salon"
	"github.com/je4/salon-digital/v2/pkg/tplfunctions"
	"github.com/je4/zsearch/v2/pkg/search"
	"github.com/op/go-logging"
	"html/template"
	"image"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type BangBang struct {
	index      bleve.Index
	urlExt     *url.URL
	dataUrl    *url.URL
	logger     *logging.Logger
	dev        bool
	collagePos map[string][]image.Rectangle
	templates  map[string]*template.Template
	templateFS fs.FS
}

func NewBangBang(index bleve.Index, urlExt *url.URL, dataUrl *url.URL, collagePos map[string][]image.Rectangle, templateFS fs.FS, logger *logging.Logger, dev bool) (*BangBang, error) {
	b := &BangBang{
		index:      index,
		urlExt:     urlExt,
		dataUrl:    dataUrl,
		collagePos: collagePos,
		logger:     logger,
		dev:        dev,
		templateFS: templateFS,
		templates:  map[string]*template.Template{},
	}
	return b, b.initTemplates()
}

func findAllFiles(fsys fs.FS, dir, suffix string) ([]string, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, emperror.Wrapf(err, "error reading directory %s", dir)
	}
	var result = []string{}
	for _, entry := range entries {
		name := filepath.ToSlash(filepath.Join(dir, entry.Name()))
		if entry.IsDir() {
			entries2, err := findAllFiles(fsys, name, suffix)
			if err != nil {
				return nil, err
			}
			result = append(result, entries2...)
		} else {
			if strings.HasSuffix(entry.Name(), suffix) {
				result = append(result, name)
			}
		}
	}
	return result, nil
}

/*
var mediaserverRegexp = regexp.MustCompile("^mediaserver:([^/]+)/([^/]+)/(.+)$")

func mediaUrl(extension, mediaserverUrl string) (string, error) {
	matches := mediaserverRegexp.FindStringSubmatch(mediaserverUrl)
	if matches == nil {
		return "", errors.New(fmt.Sprintf("invalid url: %s", mediaserverUrl))
	}
	collection := matches[1]
	signature := matches[2]
	function := matches[3]

	functions := strings.Split(strings.ToLower(function), "/")
	cmd := functions[0]
	functions = functions[1:]
	sort.Strings(functions)
	functions = append([]string{cmd}, functions...)
	function = strings.Join(functions, "/")
	filename := strings.ToLower(fmt.Sprintf("%s_%s_%s.%s",
		collection,
		strings.ReplaceAll(signature, "$", "-"),
		strings.ReplaceAll(function, "/", "_"),
		strings.TrimPrefix(extension, ".")))
	if len(filename) > 203 {
		filename = fmt.Sprintf("%s-_-%s", filename[:100], filename[len(filename)-100:])
	}
	fullpath := filepath.Join(filename)
	return fullpath, nil
}
*/
func (bb *BangBang) initTemplates() error {
	funcMap := sprig.FuncMap()
	for k, v := range tplfunctions.GetFuncMap() {
		funcMap[k] = v
	}
	templateFiles, err := findAllFiles(bb.templateFS, ".", ".gohtml")
	if err != nil {
		return emperror.Wrap(err, "cannot find templates")
	}
	for _, templateFile := range templateFiles {
		name := filepath.Base(templateFile)
		bb.templates[name], err = template.New(name).Funcs(funcMap).ParseFS(bb.templateFS, templateFile)
		if err != nil {
			return emperror.Wrapf(err, "cannot parse template: %s", templateFile)
		}
	}
	return nil

}

func (bb *BangBang) GetWork(signature string) (*search.SourceData, error) {
	raw, err := bb.index.GetInternal([]byte(signature))
	if err != nil {
		return nil, emperror.Wrapf(err, "cannot get document #%s from index", signature)
	}
	var src = &search.SourceData{}
	if err := json.Unmarshal(raw, src); err != nil {
		return nil, emperror.Wrapf(err, "cannot unmarshal document #%s", signature)
	}
	return src, nil
}
func (bb *BangBang) GetSignatureP(posX, posY int) string {
	for sig, rects := range bb.collagePos {
		for _, rect := range rects {
			if posX >= rect.Min.X && posX <= rect.Max.X {
				if posY >= rect.Min.Y && posY <= rect.Max.Y {
					return sig
				}
			}
		}
	}
	return ""
}
func (bb *BangBang) GetWorks() ([]*search.SourceData, error) {
	bQuery := bleve.NewMatchAllQuery()
	bSearch := bleve.NewSearchRequest(bQuery)
	var works = []*search.SourceData{}
	bSearch.Size = 100
	for {
		searchResult, err := bb.index.Search(bSearch)
		if err != nil {
			return nil, emperror.Wrap(err, "cannot load works from index")
		}
		for _, val := range searchResult.Hits {
			src, err := bb.GetWork(val.ID)
			if err != nil {
				return nil, emperror.Wrapf(err, "cannot get document #%s from index", val.ID)
			}
			works = append(works, src)
		}
		if len(searchResult.Hits) < 100 {
			break
		}
		bSearch.From += 100
	}
	return works, nil
}
func (bb *BangBang) GetWorksSalon() (map[string]*salon.Work, error) {
	signatures := map[string]*salon.Work{}
	works, err := bb.GetWorks()
	if err != nil {
		return nil, emperror.Wrap(err, "cannot load works from index")
	}
	for _, src := range works {
		src, err := bb.GetWork(src.Signature)
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot get document #%s from index", src.Signature)
		}
		poster := src.GetPoster()
		var work = &salon.Work{
			Signature:   src.Signature,
			Title:       src.Title,
			Year:        src.GetDate(),
			Authors:     []string{},
			Description: src.GetAbstract(),
		}
		if poster != nil {
			imagePath := fmt.Sprintf(
				"%s/data/werke/%s/derivate/%s",
				strings.TrimRight(bb.urlExt.String(), "/"),
				src.GetSignatureOriginal(),
				tplfunctions.MediaUrl(poster.Uri+"/resize/size1024x768/formatjpeg", "jpg"),
			)
			thumbPath := fmt.Sprintf(
				"%s/data/thumb/%s",
				strings.TrimRight(bb.urlExt.String(), "/"),
				tplfunctions.MediaUrl(poster.Uri+"/resize/size240x240/formatjpeg", "jpg"),
			)
			if err != nil {
				return nil, emperror.Wrapf(err, "cannot create path for %s", poster.Uri)
			}

			work.ImageUrl = imagePath
			work.ThumbnailUrl = thumbPath
		}
		iframeUrl, err := bb.urlExt.Parse(fmt.Sprintf("document/%s", src.Signature))
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot parse url %s -> document/%v", bb.urlExt.String(), src.Signature)
		}
		work.IFrameUrl = iframeUrl.String()
		for _, p := range src.GetPersons() {
			found := false
			for _, a := range work.Authors {
				if a == p.Name {
					found = true
					break
				}
			}
			if !found {
				work.Authors = append(work.Authors, p.Name)
			}
		}
		signatures[src.Signature] = work
	}
	return signatures, nil
}

func (bb *BangBang) ListHandler(w http.ResponseWriter, r *http.Request) {
	if bb.dev {
		bb.initTemplates()
	}
	tpl, ok := bb.templates["list.gohtml"]
	if !ok {
		http.Error(w, "cannot find document.gohtml", http.StatusInternalServerError)
		return
	}
	works, err := bb.GetWorks()
	if err != nil {
		bb.logger.Errorf("cannot get works: %v", err)
		http.Error(w, fmt.Sprintf("cannot get works: %v", err), http.StatusInternalServerError)
		return
	}
	salonUrl, err := bb.urlExt.Parse("/salon/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/salon"), http.StatusInternalServerError)
		return
	}
	listUrl, err := bb.urlExt.Parse("/list/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/list"), http.StatusInternalServerError)
		return
	}
	gridUrl, err := bb.urlExt.Parse("/grid/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/grid"), http.StatusInternalServerError)
		return
	}
	panoUrl, err := bb.urlExt.Parse("/pano/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/pano"), http.StatusInternalServerError)
		return
	}
	data := struct {
		Items    []*search.SourceData
		DataDir  string
		SalonUrl string
		ListUrl  string
		GridUrl  string
		PanoUrl  string
	}{
		Items:    works,
		DataDir:  bb.dataUrl.String(),
		SalonUrl: salonUrl.String(),
		ListUrl:  listUrl.String(),
		GridUrl:  gridUrl.String(),
		PanoUrl:  panoUrl.String(),
	}

	if err := tpl.Execute(w, data); err != nil {
		bb.logger.Errorf("cannot execute template: %v", err)
		http.Error(w, fmt.Sprintf("cannot execute template: %v", err), http.StatusInternalServerError)
		return
	}

}

func (bb *BangBang) GridHandler(w http.ResponseWriter, r *http.Request) {
	if bb.dev {
		bb.initTemplates()
	}
	tpl, ok := bb.templates["grid.gohtml"]
	if !ok {
		http.Error(w, "cannot find document.gohtml", http.StatusInternalServerError)
		return
	}
	works, err := bb.GetWorks()
	if err != nil {
		bb.logger.Errorf("cannot get works: %v", err)
		http.Error(w, fmt.Sprintf("cannot get works: %v", err), http.StatusInternalServerError)
		return
	}
	salonUrl, err := bb.urlExt.Parse("/salon/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/salon"), http.StatusInternalServerError)
		return
	}
	listUrl, err := bb.urlExt.Parse("/list/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/list"), http.StatusInternalServerError)
		return
	}
	gridUrl, err := bb.urlExt.Parse("/grid/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/grid"), http.StatusInternalServerError)
		return
	}
	panoUrl, err := bb.urlExt.Parse("/pano/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/pano"), http.StatusInternalServerError)
		return
	}
	data := struct {
		Items    []*search.SourceData
		DataDir  string
		SalonUrl string
		ListUrl  string
		GridUrl  string
		PanoUrl  string
	}{
		Items:    works,
		DataDir:  bb.dataUrl.String(),
		SalonUrl: salonUrl.String(),
		ListUrl:  listUrl.String(),
		GridUrl:  gridUrl.String(),
		PanoUrl:  panoUrl.String(),
	}

	if err := tpl.Execute(w, data); err != nil {
		bb.logger.Errorf("cannot execute template: %v", err)
		http.Error(w, fmt.Sprintf("cannot execute template: %v", err), http.StatusInternalServerError)
		return
	}

}

func (bb *BangBang) SalonHandler(w http.ResponseWriter, r *http.Request) {
	if bb.dev {
		bb.initTemplates()
	}
	tpl, ok := bb.templates["salon.gohtml"]
	if !ok {
		http.Error(w, "cannot find document.gohtml", http.StatusInternalServerError)
		return
	}
	salonUrl, err := bb.urlExt.Parse("/salon/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/salon"), http.StatusInternalServerError)
		return
	}
	listUrl, err := bb.urlExt.Parse("/list/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/list"), http.StatusInternalServerError)
		return
	}
	gridUrl, err := bb.urlExt.Parse("/grid/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/grid"), http.StatusInternalServerError)
		return
	}
	panoUrl, err := bb.urlExt.Parse("/pano/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/pano"), http.StatusInternalServerError)
		return
	}
	data := struct {
		DataDir  string
		PanoUrl  string
		SalonUrl string
		ListUrl  string
		GridUrl  string
	}{
		SalonUrl: salonUrl.String(),
		ListUrl:  listUrl.String(),
		GridUrl:  gridUrl.String(),
		DataDir:  bb.dataUrl.String(),
		PanoUrl:  panoUrl.String(),
	}
	if err := tpl.Execute(w, data); err != nil {
		bb.logger.Errorf("cannot execute template: %v", err)
		http.Error(w, fmt.Sprintf("cannot execute template: %v", err), http.StatusInternalServerError)
		return
	}

}

func (bb *BangBang) ZoomHandler(w http.ResponseWriter, r *http.Request) {
	if bb.dev {
		bb.initTemplates()
	}
	tpl, ok := bb.templates["zoom.gohtml"]
	if !ok {
		http.Error(w, "cannot find zoom.gohtml", http.StatusInternalServerError)
		return
	}
	salonUrl, err := bb.urlExt.Parse("/salon/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/salon"), http.StatusInternalServerError)
		return
	}
	listUrl, err := bb.urlExt.Parse("/list/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/list"), http.StatusInternalServerError)
		return
	}
	gridUrl, err := bb.urlExt.Parse("/grid/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/grid"), http.StatusInternalServerError)
		return
	}
	panoUrl, err := bb.urlExt.Parse("/pano/")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse url %s -> %s", bb.urlExt.String(), "/pano"), http.StatusInternalServerError)
		return
	}
	data := struct {
		DataDir  string
		PanoUrl  string
		SalonUrl string
		ListUrl  string
		GridUrl  string
	}{
		SalonUrl: salonUrl.String(),
		ListUrl:  listUrl.String(),
		GridUrl:  gridUrl.String(),
		DataDir:  bb.dataUrl.String(),
		PanoUrl:  panoUrl.String(),
	}
	if err := tpl.Execute(w, data); err != nil {
		bb.logger.Errorf("cannot execute template: %v", err)
		http.Error(w, fmt.Sprintf("cannot execute template: %v", err), http.StatusInternalServerError)
		return
	}

}

func (bb *BangBang) DocumentHandler(w http.ResponseWriter, r *http.Request) {
	if bb.dev {
		bb.initTemplates()
	}
	vars := mux.Vars(r)
	signature, ok := vars["signature"]
	if !ok {
		http.Error(w, "no signature in url", http.StatusNotFound)
		return
	}
	src, err := bb.GetWork(signature)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot get work #%s", signature), http.StatusNotFound)
		return
	}

	tpl, ok := bb.templates["document.gohtml"]
	if !ok {
		http.Error(w, "cannot find document.gohtml", http.StatusInternalServerError)
		return
	}
	data := struct {
		Item    *search.SourceData
		DataDir string
	}{
		Item:    src,
		DataDir: bb.dataUrl.String(),
	}
	if err := tpl.Execute(w, data); err != nil {
		bb.logger.Errorf("cannot execute template: %v", err)
		http.Error(w, fmt.Sprintf("cannot execute template: %v", err), http.StatusInternalServerError)
		return
	}
}
