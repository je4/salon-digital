package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"github.com/je4/PictureFS/v2/pkg/PictureFS"
	"github.com/je4/salon-digital/v2/pkg/bangbang"
	"github.com/je4/salon-digital/v2/pkg/browserControl"
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
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	var err error

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)

	var basedir = flag.String("basedir", ".", "base folder with html contents")
	var configfile = flag.String("cfg", filepath.Join(exPath, "salon-digital.toml"), "configuration file")

	flag.Parse()

	var config = &SalonDigitalConfig{
		LogFile:        "",
		LogLevel:       "DEBUG",
		LogFormat:      `%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}`,
		BaseDir:        *basedir,
		DataDir:        exPath,
		Addr:           "localhost:8088",
		AddrExt:        "http://localhost:8088/",
		BrowserURL:     "http://localhost:8088/digitale-see/",
		BrowserTimeout: duration{Duration: time.Minute * 5},
		User:           "",
		Password:       "",
		Browser:        true,
		Station:        true,
		BleveIndex:     filepath.Join(exPath, "bangbang.bleve"),
		Salon: SalonConfig{
			TemplateDev:    false,
			TemplateDir:    "",
			StaticDir:      "",
			PictureFSImage: "",
			PictureFSJSON:  "",
		},
		Bang: BangConfig{
			TemplateDev: false,
			TemplateDir: "",
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

	var staticFS, salonTemplateFS, bangTemplateFS fs.FS

	if config.Salon.StaticDir == "" {
		staticFS, err = fs.Sub(salon.StaticFS, "embed/static")
		if err != nil {
			logger.Panicf("cannot get subtree of static: %v", err)
		}
	} else {
		staticFS = os.DirFS(config.Salon.StaticDir)
	}

	if config.Salon.TemplateDir == "" {
		salonTemplateFS, err = fs.Sub(salon.TemplateFS, "embed/template")
		if err != nil {
			logger.Panicf("cannot get subtree of static: %v", err)
		}
	} else {
		salonTemplateFS = os.DirFS(config.Salon.TemplateDir)
	}

	index, err := bleve.Open(config.BleveIndex)
	if err != nil {
		logger.Panicf("cannot load bleve index %s: %v", config.BleveIndex, err)
	}
	defer index.Close()

	if config.Bang.TemplateDir == "" {
		bangTemplateFS, err = fs.Sub(bangbang.TemplateFS, "embed/template")
		if err != nil {
			logger.Panicf("cannot get subtree of static: %v", err)
		}
	} else {
		bangTemplateFS = os.DirFS(config.Bang.TemplateDir)
	}

	var collagePos = map[string][]image.Rectangle{}
	collageFilename := filepath.Join(config.DataDir, "collage.json")
	fp, err := os.Open(collageFilename)
	if err != nil {
		logger.Panicf("cannot open %s: %v", collageFilename, err)
	}
	jsonDec := json.NewDecoder(fp)
	if err := jsonDec.Decode(&collagePos); err != nil {
		fp.Close()
		logger.Panicf("cannot decode %s: %v", collageFilename, err)
	}
	fp.Close()

	dataUrl, err := urlExt.Parse("data/")
	if err != nil {
		logger.Panicf("cannot parse url %s -> %s: %v", urlExt.String(), "data", err)
	}
	bb, err := bangbang.NewBangBang(index, urlExt, dataUrl, collagePos, bangTemplateFS, logger, config.Station, config.Bang.TemplateDev)
	if err != nil {
		logger.Panicf("cannot instantiate bangbang: %v", err)
	}
	works, err := bb.GetWorksSalon()
	if err != nil {
		logger.Panicf("cannot get signature data: %v", err)
	}

	logger.Infof("%v works loaded", len(works))

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
		config.AddrExt,
		staticFS,
		salonTemplateFS,
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
	srv.AddSubServer("/salon-digital", salonDigital)
	bbd := &bangbang.BBDocs{BangBang: bb}
	srv.AddSubServer("/document", bbd)
	bbg := &bangbang.BBGrid{BangBang: bb}
	srv.AddSubServer("/grid", bbg)
	bbl := &bangbang.BBList{BangBang: bb}
	srv.AddSubServer("/list", bbl)
	bbs := &bangbang.BBSalon{BangBang: bb}
	srv.AddSubServer("/salon", bbs)
	bbz := &bangbang.BBZoom{BangBang: bb}
	srv.AddSubServer("/digitale-see", bbz)
	bbHome := &bangbang.BBHome{BangBang: bb}
	srv.AddSubServer("", bbHome)

	go func() {
		if err := srv.ListenAndServe(config.CertPem, config.KeyPem); err != nil {
			log.Fatalf("server died: %v", err)
		}
	}()

	var b *browserControl.BrowserControl

	if config.Browser {
		opts := map[string]any{
			"headless":                            false,
			"start-fullscreen":                    true,
			"disable-notifications":               true,
			"disable-infobars":                    true,
			"disable-gpu":                         false,
			"disable-audio-output":                false,
			"mute-audio":                          false,
			"allow-insecure-localhost":            true,
			"enable-immersive-fullscreen-toolbar": true,
			"views-browser-windows":               false,
			"kiosk":                               true,
			"disable-session-crashed-bubble":      true,
			"incognito":                           true,
			//				"enable-features":                     "PreloadMediaEngagementData,AutoplayIgnoreWebAudio,MediaEngagementBypassAutoplayPolicies",
			//			"disable-features": "InfiniteSessionRestore,TranslateUI,PreloadMediaEngagementData,AutoplayIgnoreWebAudio,MediaEngagementBypassAutoplayPolicies",
			"disable-features": "InfiniteSessionRestore,TranslateUI,PreloadMediaEngagementData,AutoplayIgnoreWebAudio,MediaEngagementBypassAutoplayPolicies",
			//"no-first-run":                        true,
			"enable-fullscreen-toolbar-reveal": false,
			"useAutomationExtension":           false,
			"enable-automation":                false,
		}
		homeUrl, err := url.Parse(config.BrowserURL)
		if err != nil {
			logger.Panicf("cannot parse %s: %v", config.AddrExt, err)
		}
		b, err = browserControl.NewBrowserControl(config.AddrExt, homeUrl, opts, config.BrowserTimeout.Duration, logger)
		if err != nil {
			logger.Panicf("cannot create browser control: %v", err)
		}
		b.Start()
	}

	end := make(chan bool, 1)

	// process waiting for interrupt signal (TERM or KILL)
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)

		signal.Notify(sigint, syscall.SIGTERM)
		signal.Notify(sigint, syscall.SIGKILL)

		<-sigint

		if config.Browser && b != nil {
			b.Shutdown()
		}

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
