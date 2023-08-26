package bangbang

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/gorilla/mux"
	"github.com/je4/salon-digital/v2/pkg/salon"
	"github.com/je4/salon-digital/v2/pkg/tplfunctions"
	"github.com/je4/zsearch/v2/pkg/search"
	"golang.org/x/exp/slices"
	"html/template"
	"image"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type stats struct {
	Items, Images, PDFs, Videos, VideoLength, Audios, AudioLength, Persons int64
}

type BangBang struct {
	index      bleve.Index
	urlExt     *url.URL
	dataUrl    *url.URL
	logger     *logging.Logger
	dev        bool
	station    bool
	collagePos map[string][]image.Rectangle
	templates  map[string]*template.Template
	templateFS fs.FS
	panoUrl    *url.URL
	salonUrl   *url.URL
	listUrl    *url.URL
	gridUrl    *url.URL
	salonZoom  float64
	works      []*search.SourceData
	stats      stats
}

func NewBangBang(index bleve.Index, urlExt *url.URL, dataUrl *url.URL, collagePos map[string][]image.Rectangle, templateFS fs.FS, salonZoom float64, logger *logging.Logger, station, dev bool) (*BangBang, error) {
	b := &BangBang{
		index:      index,
		urlExt:     urlExt,
		dataUrl:    dataUrl,
		collagePos: collagePos,
		logger:     logger,
		station:    station,
		dev:        dev,
		templateFS: templateFS,
		templates:  map[string]*template.Template{},
		salonZoom:  salonZoom,
	}
	var err error
	b.salonUrl, err = b.urlExt.Parse("salon/")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse url %s -> %s", b.urlExt.String(), "/salon")
	}
	b.listUrl, err = b.urlExt.Parse("list/")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse url %s -> %s", b.urlExt.String(), "/list")
	}
	b.gridUrl, err = b.urlExt.Parse("grid/")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse url %s -> %s", b.urlExt.String(), "/grid")
	}
	b.panoUrl, err = b.urlExt.Parse("digitale-see/")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse url %s -> %s", b.urlExt.String(), "/pano")
	}

	return b, b.initTemplates()
}

func findAllFiles(fsys fs.FS, dir, suffix string) ([]string, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading directory %s", dir)
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
		return errors.Wrap(err, "cannot find templates")
	}
	for _, templateFile := range templateFiles {
		name := filepath.Base(templateFile)
		bb.templates[name], err = template.New(name).Funcs(funcMap).ParseFS(bb.templateFS, templateFile)
		if err != nil {
			return errors.Wrapf(err, "cannot parse template: %s", templateFile)
		}
	}
	return nil

}

func (bb *BangBang) GetWork(signature string) (*search.SourceData, error) {
	raw, err := bb.index.GetInternal([]byte(signature))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get document #%s from index", signature)
	}
	var src = &search.SourceData{}
	if err := json.Unmarshal(raw, src); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal document #%s", signature)
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
	if bb.works != nil {
		return bb.works, nil
	}
	bQuery := bleve.NewMatchAllQuery()
	bSearch := bleve.NewSearchRequest(bQuery)
	bb.works = []*search.SourceData{}
	bSearch.Size = 100
	for {
		searchResult, err := bb.index.Search(bSearch)
		if err != nil {
			return nil, errors.Wrap(err, "cannot load works from index")
		}
		for _, val := range searchResult.Hits {
			src, err := bb.GetWork(val.ID)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot get document #%s from index", val.ID)
			}
			bb.works = append(bb.works, src)
		}
		if len(searchResult.Hits) < 100 {
			break
		}
		bSearch.From += 100
	}

	var pstring = []string{}
	bb.stats.Items = int64(len(bb.works))
	for _, item := range bb.works {
		for t, ms := range item.GetMedia() {
			switch t {
			case "image":
				bb.stats.Images += int64(len(ms))
			case "pdf":
				bb.stats.PDFs += int64(len(ms))
			case "video":
				bb.stats.Videos += int64(len(ms))
				for _, m := range ms {
					bb.stats.VideoLength += m.Duration
				}
			case "audio":
				bb.stats.Audios += int64(len(ms))
				for _, m := range ms {
					bb.stats.AudioLength += m.Duration
				}
			}
			badname := regexp.MustCompile("^[^,:]+(,[^,:]+)?$")
			for _, p := range item.GetPersons() {
				name := strings.TrimSpace(strings.ToLower(p.Name))
				if name == "" {
					continue
				}
				if !badname.MatchString(name) {
					fmt.Printf("#%s: [%s] %s\n", item.GetSignatureOriginal(), p.Role, p.Name)
					continue
				}
				found := false
				for _, s := range pstring {
					//					d := tdl.Distance(name, s)
					//					fmt.Printf("%s <-> %s: %v;\n", name, s, d)
					if s == name {
						found = true
						break
					}
				}
				if !found {
					pstring = append(pstring, name)
					// fmt.Printf("%v: %s\n", len(pstring), name)
				}
			}
			bb.stats.Persons = int64(len(pstring))
		}
	}
	fmt.Printf("Items: %v\n", bb.stats.Items)
	fmt.Printf("Images: %v\n", bb.stats.Images)
	fmt.Printf("PDFs: %v\n", bb.stats.PDFs)
	fmt.Printf("Videos: %v (%vsec)\n", bb.stats.Videos, bb.stats.VideoLength)
	fmt.Printf("Audios: %v (%vsec)\n", bb.stats.Audios, bb.stats.AudioLength)

	return bb.works, nil
}
func (bb *BangBang) GetWorksSalon() (map[string]*salon.Work, error) {
	signatures := map[string]*salon.Work{}
	works, err := bb.GetWorks()
	if err != nil {
		return nil, errors.Wrap(err, "cannot load works from index")
	}
	for _, src := range works {
		src, err := bb.GetWork(src.Signature)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get document #%s from index", src.Signature)
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
				return nil, errors.Wrapf(err, "cannot create path for %s", poster.Uri)
			}

			work.ImageUrl = imagePath
			work.ThumbnailUrl = thumbPath
		}
		iframeUrl, err := bb.urlExt.Parse(fmt.Sprintf("document/%s", src.Signature))
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse url %s -> document/%v", bb.urlExt.String(), src.Signature)
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
	data := struct {
		Items    []*search.SourceData
		DataDir  string
		SalonUrl string
		ListUrl  string
		GridUrl  string
		PanoUrl  string
		Station  bool
	}{
		Items:    works,
		DataDir:  bb.dataUrl.String(),
		SalonUrl: bb.salonUrl.String(),
		ListUrl:  bb.listUrl.String(),
		GridUrl:  bb.gridUrl.String(),
		PanoUrl:  bb.panoUrl.String(),
		Station:  bb.station,
	}

	if err := tpl.Execute(w, data); err != nil {
		bb.logger.Errorf("cannot execute template: %v", err)
		http.Error(w, fmt.Sprintf("cannot execute template: %v", err), http.StatusInternalServerError)
		return
	}

}

func (bb *BangBang) GridHandler(w http.ResponseWriter, r *http.Request) {
	restrict := r.URL.Query().Get("restrict")
	if bb.dev {
		bb.initTemplates()
	}
	var tplName string
	if restrict == "bangbang" {
		tplName = "gridbang.gohtml"
	} else {
		tplName = "grid.gohtml"
	}
	tpl, ok := bb.templates[tplName]
	if !ok {
		http.Error(w, fmt.Sprintf("cannot find %s", tplName), http.StatusInternalServerError)
		return
	}
	works, err := bb.GetWorks()
	if err != nil {
		bb.logger.Errorf("cannot get works: %v", err)
		http.Error(w, fmt.Sprintf("cannot get works: %v", err), http.StatusInternalServerError)
		return
	}
	var items []*search.SourceData
	if restrict == "bangbang" {
		items = []*search.SourceData{}
		for _, item := range works {
			if strings.HasPrefix(item.GetTitle(), "BANG BANG:") {
				items = append(items, item)
			}
		}
	} else {
		items = works
	}
	data := struct {
		Items    []*search.SourceData
		DataDir  string
		SalonUrl string
		ListUrl  string
		GridUrl  string
		PanoUrl  string
		Station  bool
	}{
		Items:    items,
		DataDir:  bb.dataUrl.String(),
		SalonUrl: bb.salonUrl.String(),
		ListUrl:  bb.listUrl.String(),
		GridUrl:  bb.gridUrl.String(),
		PanoUrl:  bb.panoUrl.String(),
		Station:  bb.station,
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
	data := struct {
		DataDir   string
		PanoUrl   string
		SalonUrl  string
		ListUrl   string
		GridUrl   string
		Station   bool
		SalonZoom float64
	}{
		SalonUrl:  bb.salonUrl.String(),
		ListUrl:   bb.listUrl.String(),
		GridUrl:   bb.gridUrl.String(),
		DataDir:   bb.dataUrl.String(),
		PanoUrl:   bb.panoUrl.String(),
		Station:   bb.station,
		SalonZoom: bb.salonZoom,
	}
	if err := tpl.Execute(w, data); err != nil {
		bb.logger.Errorf("cannot execute template: %v", err)
		http.Error(w, fmt.Sprintf("cannot execute template: %v", err), http.StatusInternalServerError)
		return
	}

}

func (bb *BangBang) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if bb.dev {
		bb.initTemplates()
	}
	tpl, ok := bb.templates["home.gohtml"]
	if !ok {
		http.Error(w, "cannot find document.gohtml", http.StatusInternalServerError)
		return
	}
	lang := strings.ToLower(r.URL.Query().Get("lang"))
	if lang != "en" && lang != "fr" {
		lang = "de"
	}
	data := struct {
		Lang     string
		DataDir  string
		PanoUrl  string
		SalonUrl string
		ListUrl  string
		GridUrl  string
		Station  bool
		Stats    stats
	}{
		Lang:     lang,
		SalonUrl: bb.salonUrl.String(),
		ListUrl:  bb.listUrl.String(),
		GridUrl:  bb.gridUrl.String(),
		DataDir:  bb.dataUrl.String(),
		PanoUrl:  bb.panoUrl.String(),
		Station:  bb.station,
		Stats:    bb.stats,
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
	data := struct {
		DataDir  string
		PanoUrl  string
		SalonUrl string
		ListUrl  string
		GridUrl  string
		Station  bool
	}{
		SalonUrl: bb.salonUrl.String(),
		ListUrl:  bb.listUrl.String(),
		GridUrl:  bb.gridUrl.String(),
		DataDir:  bb.dataUrl.String(),
		PanoUrl:  bb.panoUrl.String(),
		Station:  bb.station,
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
	contentOnline := false
	acl_content, ok := src.ACL["content"]
	if ok {
		contentOnline = len(acl_content) > 1
		for _, acl := range acl_content {
			if acl == "global/guest" {
				contentOnline = true
			}
		}
	}

	tpl, ok := bb.templates["document.gohtml"]
	if !ok {
		http.Error(w, "cannot find document.gohtml", http.StatusInternalServerError)
		return
	}
	data := struct {
		Item          *search.SourceData
		DataDir       string
		Station       bool
		ContentOnline bool
	}{
		Item:          src,
		DataDir:       bb.dataUrl.String(),
		Station:       bb.station,
		ContentOnline: contentOnline,
	}
	if err := tpl.Execute(w, data); err != nil {
		bb.logger.Errorf("cannot execute template: %v", err)
		http.Error(w, fmt.Sprintf("cannot execute template: %v", err), http.StatusInternalServerError)
		return
	}
}

func (bb *BangBang) DetailHandler(w http.ResponseWriter, r *http.Request) {
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
	contentOnline := false
	acl_content, ok := src.ACL["content"]
	if ok {
		contentOnline = len(acl_content) > 1
		for _, acl := range acl_content {
			if acl == "global/guest" {
				contentOnline = true
			}
		}
	}

	tpl, ok := bb.templates["detail.gohtml"]
	if !ok {
		http.Error(w, "cannot find document.gohtml", http.StatusInternalServerError)
		return
	}
	data := struct {
		Item          *search.SourceData
		DataDir       string
		Station       bool
		ContentOnline bool
		SalonUrl      string
		ListUrl       string
		GridUrl       string
		PanoUrl       string
	}{
		Item:          src,
		DataDir:       bb.dataUrl.String(),
		Station:       bb.station,
		ContentOnline: contentOnline,
		SalonUrl:      bb.salonUrl.String(),
		ListUrl:       bb.listUrl.String(),
		GridUrl:       bb.gridUrl.String(),
		PanoUrl:       bb.panoUrl.String(),
	}
	if err := tpl.Execute(w, data); err != nil {
		bb.logger.Errorf("cannot execute template: %v", err)
		http.Error(w, fmt.Sprintf("cannot execute template: %v", err), http.StatusInternalServerError)
		return
	}
}

type personIndex struct {
	Name  string
	Works map[string][]string
}

func (bb *BangBang) IndexHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if bb.dev {
		bb.initTemplates()
	}
	var tplName = "index.gohtml"
	tpl, ok := bb.templates[tplName]
	if !ok {
		http.Error(w, fmt.Sprintf("cannot find %s", tplName), http.StatusInternalServerError)
		return
	}
	works, err := bb.GetWorks()
	if err != nil {
		bb.logger.Errorf("cannot get works: %v", err)
		http.Error(w, fmt.Sprintf("cannot get works: %v", err), http.StatusInternalServerError)
		return
	}
	sort.Slice(works, func(i, j int) bool {
		a := works[i]
		b := works[j]
		if len(a.Signature) < len(b.Signature) {
			return true
		}
		return a.Signature < b.Signature
	})

	pIndex := map[string]personIndex{}
	pNames := []string{}
	badname := regexp.MustCompile("^[^,:]+(,[^,:]+)?$")
	roles := []string{"artist", "performer", "director"}
	for _, item := range works {
		for _, p := range item.GetPersons() {
			name := strings.Trim(strings.ToLower(p.Name), " [];,")
			parts2 := strings.Split(name, "(")
			name = strings.TrimSpace(parts2[0])
			if name == "" {
				continue
			}
			if !slices.Contains(roles, p.Role) {
				continue
			}
			if p.Role == "director" {
				p.Role = "artist"
			}
			if !badname.MatchString(name) {
				fmt.Printf("#%s: [%s] %s\n", item.GetSignatureOriginal(), p.Role, p.Name)
				//continue
			}
			parts := strings.Split(p.Name, "(")
			if _, ok := pIndex[name]; !ok {
				pIndex[name] = personIndex{
					Name:  strings.TrimSpace(parts[0]),
					Works: map[string][]string{},
				}
				pNames = append(pNames, name)
			}
			if _, ok := pIndex[name].Works[p.Role]; !ok {
				pIndex[name].Works[p.Role] = []string{}
			}
			id := item.GetSignatureOriginal()
			for {
				if len(id) > 4 {
					break
				}
				id = "0" + id
			}
			id = "#" + id
			pIndex[name].Works[p.Role] = append(pIndex[name].Works[p.Role], id)
		}
	}
	sort.Strings(pNames)
	data := struct {
		Items       []*search.SourceData
		DataDir     string
		SalonUrl    string
		ListUrl     string
		GridUrl     string
		PanoUrl     string
		Station     bool
		Persons     map[string]personIndex
		PersonNames []string
		Type        string
	}{
		Items:       works,
		DataDir:     bb.dataUrl.String(),
		SalonUrl:    bb.salonUrl.String(),
		ListUrl:     bb.listUrl.String(),
		GridUrl:     bb.gridUrl.String(),
		PanoUrl:     bb.panoUrl.String(),
		Station:     bb.station,
		Persons:     pIndex,
		PersonNames: pNames,
		Type:        vars["type"],
	}

	if err := tpl.Execute(w, data); err != nil {
		bb.logger.Errorf("cannot execute template: %v", err)
		http.Error(w, fmt.Sprintf("cannot execute template: %v", err), http.StatusInternalServerError)
		return
	}

}
