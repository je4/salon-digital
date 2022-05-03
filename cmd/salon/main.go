package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"github.com/blevesearch/bleve/v2"
	"github.com/je4/PictureFS/v2/pkg/PictureFS"
	"github.com/je4/salon-digital/v2/pkg/bangbang"
	"github.com/je4/salon-digital/v2/pkg/salon"
	"github.com/je4/salon-digital/v2/pkg/server"
	lm "github.com/je4/utils/v2/pkg/logger"
	"image"
	"io"
	"io/fs"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var err error

	var basedir = flag.String("basedir", ".", "base folder with html contents")
	var configfile = flag.String("cfg", "/etc/tbbs.toml", "configuration file")

	flag.Parse()

	var config = &SalonDigitalConfig{
		LogFile:    "",
		LogLevel:   "DEBUG",
		LogFormat:  `%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}`,
		BaseDir:    *basedir,
		Addr:       "localhost:80",
		AddrExt:    "http://localhost:80/",
		User:       "jane",
		Password:   "doe",
		BleveIndex: "",
		Salon: SalonConfig{
			TemplateDev:    false,
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

	urlExt, err := url.Parse(config.AddrExt)
	if err != nil {
		logger.Panicf("invalid addrext %s: %v", config.AddrExt, err)
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

	index, err := bleve.Open(config.BleveIndex)
	if err != nil {
		logger.Panicf("cannot load bleve index %s: %v", config.BleveIndex, err)
	}
	defer index.Close()

	dataUrl, err := urlExt.Parse("data/")
	if err != nil {
		logger.Panicf("cannot parse url %s -> %s: %v", urlExt.String(), "data", err)
	}
	bb, err := bangbang.NewBangBang(index, urlExt, dataUrl, logger)
	if err != nil {
		logger.Panicf("cannot instantiate bangbang: %v", err)
	}
	works, err := bb.GetWorks()
	if err != nil {
		logger.Panicf("cannot get signature data: %v", err)
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
		works,
		staticFS,
		templateFS,
		config.Salon.TemplateDev,
		pfs,
		logger,
	)
	if err != nil {
		logger.Panicf("cannot create salon: %v", err)
	}

	dataFS := os.DirFS(config.DataDir)

	srv, err := server.NewServer(
		"Salon-Digital",
		config.Addr,
		urlExt,
		config.User,
		config.Password,
		dataFS,
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
