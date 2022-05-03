package bangbang

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"github.com/goph/emperror"
	"github.com/je4/salon-digital/v2/pkg/salon"
	"github.com/je4/zsearch/v2/pkg/search"
	"github.com/op/go-logging"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type BangBang struct {
	index   bleve.Index
	urlExt  *url.URL
	dataUrl *url.URL
	logger  *logging.Logger
}

func NewBangBang(index bleve.Index, urlExt *url.URL, dataUrl *url.URL, logger *logging.Logger) (*BangBang, error) {
	b := &BangBang{
		index:   index,
		urlExt:  urlExt,
		dataUrl: dataUrl,
		logger:  logger,
	}
	return b, nil
}

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

func (bb *BangBang) GetWorks() (map[string]*salon.Work, error) {
	bQuery := bleve.NewMatchAllQuery()
	bSearch := bleve.NewSearchRequest(bQuery)
	searchResult, err := bb.index.Search(bSearch)
	if err != nil {
		return nil, emperror.Wrap(err, "cannot load works from index")
	}
	var signatures = map[string]*salon.Work{}
	for _, val := range searchResult.Hits {
		raw, err := bb.index.GetInternal([]byte(val.ID))
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot get document #%s from index", val.ID)
		}
		var src = &search.SourceData{}
		if err := json.Unmarshal(raw, src); err != nil {
			return nil, emperror.Wrapf(err, "cannot unmarshal document #%s", val.ID)
		}
		//logger.Info(string(raw))
		poster := src.GetPoster()
		workid, err := strconv.ParseInt(src.GetSignatureOriginal(), 10, 64)
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot convert original id %s of %f to int", src.GetSignatureOriginal(), src.GetSignature())
		}
		imagePath, err := mediaUrl(
			"jpg",
			poster.Uri+"/resize/size1024x768/formatjpeg")
		thumbPath, err := mediaUrl(
			"jpg",
			poster.Uri+"/resize/size240x240/formatjpeg")
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot create path for %s", poster.Uri)
		}

		imageUrl, err := bb.dataUrl.Parse(fmt.Sprintf("werke/%d/derivate/%s", workid, imagePath))
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot parse url %s -> %s", bb.urlExt.String(), imagePath)
		}
		thumbUrl, err := bb.dataUrl.Parse(fmt.Sprintf("werke/%d/derivate/%s", workid, thumbPath))
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot parse url %s -> %s", bb.urlExt.String(), imagePath)
		}
		iframeUrl, err := bb.urlExt.Parse(fmt.Sprintf("document/%s", val.ID))
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot parse url %s -> %s", bb.urlExt.String(), imagePath)
		}
		var work = &salon.Work{
			Signature:    val.ID,
			Title:        src.Title,
			Year:         src.GetDate(),
			Authors:      []string{},
			Description:  src.GetAbstract(),
			ImageUrl:     imageUrl.String(),
			ThumbnailUrl: thumbUrl.String(),
			IFrameUrl:    iframeUrl.String(),
		}
		signatures[val.ID] = work

	}
	return signatures, nil
}
