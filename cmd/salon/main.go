package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"github.com/je4/PictureFS/v2/pkg/PictureFS"
	"github.com/je4/salon-digital/v2/pkg/salon"
	"github.com/je4/salon-digital/v2/pkg/server"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/je4/zsearch/v2/pkg/search"
	"github.com/pkg/errors"
	"image"
	"io"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var mediaserverRegexp = regexp.MustCompile("^mediaserver:([^/]+)/([^/]+)/(.+)$")

func mediaUrl(exportPath, folder, extension, mediaserverUrl string) (string, error) {
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
	fullpath := filepath.Join(exportPath, folder, filename)
	return fullpath, nil
}

func main() {
	var err error

	var basedir = flag.String("basedir", ".", "base folder with html contents")
	var configfile = flag.String("cfg", "/etc/tbbs.toml", "configuration file")

	flag.Parse()

	var config = &SalonDigitalConfig{
		LogFile:   "",
		LogLevel:  "DEBUG",
		LogFormat: `%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}`,
		BaseDir:   *basedir,
		Addr:      "localhost:80",
		AddrExt:   "http://localhost:80/",
		User:      "jane",
		Password:  "doe",
		Salon: SalonConfig{
			TemplateDev:    false,
			BleveIndex:     "",
			TemplateDir:    "",
			StaticDir:      "",
			PictureFSImage: "",
			PictureFSJSON:  "",
		},
	}
	if err := LoadSalonDigitalConfig(*configfile, config); err != nil {
		log.Printf("cannot load config file: %v", err)
	}

	// create logger instance
	logger, lf := lm.CreateLogger("Salon Digital", config.LogFile, nil, config.LogLevel, config.LogFormat)
	defer lf.Close()

	var accessLog io.Writer
	var f *os.File
	if config.AccessLog == "" {
		accessLog = os.Stdout
	} else {
		f, err = os.OpenFile(config.AccessLog, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			logger.Panicf("cannot open file %s: %v", config.AccessLog, err)
			return
		}
		defer f.Close()
		accessLog = f
	}
	var staticFS, templateFS fs.FS

	if config.Salon.StaticDir == "" {
		staticFS, err = fs.Sub(salon.StaticFS, "static")
		if err != nil {
			logger.Panicf("cannot get subtree of static: %v", err)
		}
	} else {
		staticFS = os.DirFS(config.Salon.StaticDir)
	}

	if config.Salon.TemplateDir == "" {
		templateFS, err = fs.Sub(salon.TemplateFS, "embed/template")
		if err != nil {
			logger.Panicf("cannot get subtree of static: %v", err)
		}
	} else {
		templateFS = os.DirFS(config.Salon.TemplateDir)
	}

	index, err := bleve.Open(config.Salon.BleveIndex)
	if err != nil {
		logger.Panicf("cannot load bleve index %s: %v", config.Salon.BleveIndex, err)
	}
	defer index.Close()

	bQuery := bleve.NewMatchAllQuery()
	bSearch := bleve.NewSearchRequest(bQuery)
	searchResult, err := index.Search(bSearch)
	if err != nil {
		logger.Panicf("cannot load works from index: %v", err)
	}
	var signatures = map[string]salon.Work{}
	for _, val := range searchResult.Hits {
		raw, err := index.GetInternal([]byte(val.ID))
		if err != nil {
			logger.Panicf("cannot get document #%s from index: %v", val.ID, err)
		}
		var src = &search.SourceData{}
		if err := json.Unmarshal(raw, src); err != nil {
			logger.Panicf("cannot unmarshal document #%s: %v", val.ID, err)
		}
		//logger.Info(string(raw))
		poster := src.GetPoster()
		workid, err := strconv.ParseInt(src.GetSignatureOriginal(), 10, 64)
		if err != nil {
			logger.Panicf("cannot convert original id %s of %f to int: %v", src.GetSignatureOriginal(), src.GetSignature(), err)
		}
		imagePath, err := mediaUrl(
			config.Salon.ExportPath,
			fmt.Sprintf("werke/%d/derivate", workid),
			"jpg",
			poster.Uri+"/resize/size1024x768/formatjpeg")
		thumbPath, err := mediaUrl(
			config.Salon.ExportPath,
			fmt.Sprintf("werke/%d/derivate", workid),
			"jpg",
			poster.Uri+"/resize/size240x240/formatjpeg")
		if err != nil {
			logger.Panicf("cannot create path for %s: %v", poster.Uri, err)
		}

		var work = salon.Work{
			Signature:    val.ID,
			Title:        src.Title,
			Year:         src.GetDate(),
			Authors:      []string{},
			Description:  src.GetAbstract(),
			ImageUrl:     "file://" + filepath.ToSlash(imagePath),
			ThumbnailUrl: "file://" + filepath.ToSlash(thumbPath),
			IFrameUrl:    fmt.Sprintf("%s/document/%s", config.AddrExt, val.ID),
		}
		signatures[val.ID] = work
	}

	logger.Info("loading PictureFS...")
	var pfs *PictureFS.FS
	if config.Salon.PictureFSImage != "" {
		pfs, err = PictureFS.NewFSFile(config.Salon.PictureFSImage, config.Salon.PictureFSJSON)
		if err != nil {
			logger.Panicf("cannot load PictureFS(%s/%s): %v", config.Salon.PictureFSImage, config.Salon.PictureFSJSON, err)
		}
	} else {
		pfsImage, _, err := image.Decode(bytes.NewBuffer(salon.SalonDigitalImage))
		if err != nil {
			logger.Panicf("cannot decode embedded PictureFS: %v", err)
		}

		var layout = PictureFS.Layout{}
		if err := json.Unmarshal(salon.SalonDigitalJSON, &layout); err != nil {
			logger.Panicf("cannot unmarshal embedded PictureFS: %v", err)
		}
		pfs, err = PictureFS.NewFS(pfsImage, layout)
		if err != nil {
			logger.Panicf("cannot initiate PictureFS: %v", err)
		}
	}

	salonDigital, err := salon.NewSalon(
		signatures,
		staticFS,
		templateFS,
		config.Salon.TemplateDev,
		pfs,
		logger,
	)
	if err != nil {
		logger.Panicf("cannot create salon: %v", err)
	}

	srv, err := server.NewServer(
		"Salon-Digital",
		config.Addr,
		config.AddrExt,
		config.User,
		config.Password,
		logger,
		accessLog,
	)
	if err != nil {
		logger.Panicf("cannot initialize server: %v", err)
	}
	srv.AddSubServer("/salon", salonDigital)

	go func() {
		if err := srv.ListenAndServe(config.CertPem, config.KeyPem); err != nil {
			log.Fatalf("server died: %v", err)
		}
	}()

	end := make(chan bool, 1)

	// process waiting for interrupt signal (TERM or KILL)
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)

		signal.Notify(sigint, syscall.SIGTERM)
		signal.Notify(sigint, syscall.SIGKILL)

		<-sigint

		// We received an interrupt signal, shut down.
		logger.Infof("shutdown requested")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		srv.Shutdown(ctx)

		end <- true
	}()

	<-end
	logger.Info("server stopped")

}
