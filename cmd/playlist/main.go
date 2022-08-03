package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"github.com/je4/salon-digital/v2/pkg/tplfunctions"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/je4/zsearch/v2/pkg/search"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type entry struct {
	VideoFile string
	Title     string
	Author    string
	Signature string
	Year      string
	Len       int64
}

func timeMustParse(layout, value string) time.Time {
	result, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return result
}

func inDateList(t time.Time, tl []time.Time) bool {
	for _, te := range tl {
		if t == te {
			return true
		}
	}
	return false
}

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
		LogFile:             "",
		LogLevel:            "DEBUG",
		LogFormat:           `%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}`,
		BaseDir:             *basedir,
		DataDir:             exPath,
		Addr:                "localhost:8088",
		AddrExt:             "http://localhost:8088/",
		BrowserURL:          "http://localhost:8088/digitale-see/",
		BrowserTimeout:      duration{Duration: time.Minute * 5},
		User:                "",
		Password:            "",
		Browser:             true,
		BrowserTaskDelay:    duration{Duration: time.Second * 2},
		Station:             true,
		StationStartTimeout: duration{Duration: time.Second * 20},
		BleveIndex:          filepath.Join(exPath, "bangbang.bleve"),
		Salon: SalonConfig{
			TemplateDev:    false,
			TemplateDir:    "",
			StaticDir:      "",
			PictureFSImage: "",
			PictureFSJSON:  "",
			Zoom:           1.0,
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

	index, err := bleve.Open(config.BleveIndex)
	if err != nil {
		logger.Panicf("cannot load bleve index %s: %v", config.BleveIndex, err)
	}
	defer index.Close()

	var list = []*entry{}
	bQuery := bleve.NewMatchAllQuery()
	bSearch := bleve.NewSearchRequest(bQuery)
	bSearch.Size = 100
	items := 0
	for {
		searchResult, err := index.Search(bSearch)
		if err != nil {
			logger.Panicf("cannot load works from index: %v", err)
		}
		for _, val := range searchResult.Hits {
			raw, err := index.GetInternal([]byte(val.ID))
			if err != nil {
				logger.Panicf("cannot get document #%s from index: %v", val.ID, err)
			}
			var src = &search.SourceData{}
			if err := json.Unmarshal(raw, src); err != nil {
				logger.Panicf("cannot unmarshal document #%s: %v", val.ID, err)
			}
			items++
			if strings.HasPrefix(src.Signature, "zotero2-") || strings.HasPrefix(src.Title, "BANG BANG:") {
				continue
			}
			for t, ms := range src.GetMedia() {
				if t != "video" {
					continue
				}
				for _, m := range ms {
					if m.Type != "video" {
						continue
					}
					id := src.GetSignatureOriginal()
					for {
						if len(id) < 4 {
							id = "0" + id
						} else {
							break
						}
					}
					fp := filepath.ToSlash(filepath.Join( /*config.DataDir, */ "werke", fmt.Sprintf("%s", src.GetSignatureOriginal()), "derivate", tplfunctions.MediaUrl(m.Uri+"$$web/master", ".mp4")))
					_, err := os.Stat(filepath.Join(config.DataDir, fp))
					if err != nil {
						logger.Errorf("#%s - cannot stat file %s", id, fp)
						continue
					}
					if m.Duration > 3600*4 {
						logger.Infof("#%s - too long", id)
						continue
					}
					entry := &entry{
						VideoFile: fp,
						Title:     fmt.Sprintf("#%s - %s", id, src.GetTitle()),
						Author:    "",
						Signature: src.GetSignature(),
						Year:      src.GetDate(),
						Len:       m.Duration,
					}
					for _, p := range src.GetPersons() {
						if strings.ToLower(p.Role) != "artist" {
							continue
						}
						entry.Author += fmt.Sprintf("; %s", p.Name)
					}
					entry.Author = strings.Trim(entry.Author, "; ")
					list = append(list, entry)
				}
			}
		}
		if len(searchResult.Hits) < 100 {
			break
		}
		bSearch.From += 100
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })

	var seconds int64 = 0

	noKino := []time.Time{timeMustParse("2006-01-02", "2022-08-12")}
	till21h := []time.Time{
		timeMustParse("2006-01-02", "2022-08-05"),
		timeMustParse("2006-01-02", "2022-08-12"),
		timeMustParse("2006-01-02", "2022-08-19")}
	till22h := []time.Time{
		timeMustParse("2006-01-02", "2022-08-13"),
		timeMustParse("2006-01-02", "2022-08-20")}
	lastDay, err := time.Parse("2006-01-02" /* "2022-08-21" */, "2022-08-19")
	if err != nil {
		logger.Panic("invalid date")
	}
	currentDay := lastDay.Add(time.Hour * 24)
	var currentTime, dayEnd time.Duration
	nextDay := func() {
		currentDay = currentDay.Add(-time.Hour * 24)
		for {
			if currentDay.Weekday() == 1 {
				currentDay = currentDay.Add(-time.Hour * 24)
				continue
			}
			if inDateList(currentDay, noKino) {
				currentDay = currentDay.Add(-time.Hour * 24)
				continue
			}
			break
		}

		currentTime = time.Hour * 11
		dayEnd = time.Hour * 18
		if currentDay.Weekday() == 4 {
			dayEnd = time.Hour * 21
		}
		if inDateList(currentDay, till21h) {
			dayEnd = time.Hour * 21
		}
		if inDateList(currentDay, till22h) {
			dayEnd = time.Hour * 22
		}
		dayEnd -= time.Minute * 15
	}

	nextDay()

	prghtml := filepath.Join(config.BaseDir, fmt.Sprintf("program.html"))
	prg2, err := os.Create(prghtml)
	if err != nil {
		logger.Panicf("cannot create playlist %s", prghtml)
	}
	prg2.WriteString("<html><head></head><body>\n")

	var dates = map[string]string{}
	pls := []*entry{}
	for key, e := range list {
		seconds += e.Len
		pls = append(pls, e)
		currentTime += time.Duration(e.Len) * time.Second
		logger.Infof("%000d #%v: [%vsec] %s - %s", key, e.Signature, e.Len, e.VideoFile, e.Title)
		if currentTime > dayEnd {
			currentTime = time.Hour * 11
			vlcname := filepath.Join(config.BaseDir, fmt.Sprintf("playlist_%s.vlc", currentDay.Format("2006-01-02")))
			vlc, err := os.Create(vlcname)
			if err != nil {
				logger.Panicf("cannot create playlist %s", vlcname)
			}
			prgname := filepath.Join(config.BaseDir, fmt.Sprintf("program_%s.txt", currentDay.Format("2006-01-02")))
			prg, err := os.Create(prgname)
			if err != nil {
				logger.Panicf("cannot create playlist %s", prgname)
			}
			prg2.WriteString(fmt.Sprintf("<span id=\"%s\"><h3>%s</h3></span>\n", currentDay.Format("20060102"), currentDay.Format("02.01.2006")))
			prg2.WriteString("<table>\n")
			dates[currentDay.Format("20060102")] = currentDay.Format("02.01.2006")

			vlc.WriteString("[playlist]\n")
			vlc.WriteString(fmt.Sprintf("NumberOfEntries=%v\n", len(pls)))

			prg.WriteString(fmt.Sprintf("Programm vom %s\n\n", currentDay.Format("02.01.2006")))
			for fno, e := range pls {
				vlc.WriteString(fmt.Sprintf("File%v=%s\n", fno+1, e.VideoFile))
				vlc.WriteString(fmt.Sprintf("Title%v=%s\n", fno+1, e.Title))

				startTime := currentDay.Add(currentTime)

				prg.WriteString(fmt.Sprintf("%s - %s\n", startTime.Format("15:04:05"), e.Title))

				prg2.WriteString("  <tr>\n")
				prg2.WriteString(fmt.Sprintf("      <td>%s</td>\n", startTime.Format("15:04:05")))
				prg2.WriteString(fmt.Sprintf("      <td>%s</td>\n", e.Title))
				prg2.WriteString("  </tr>\n")

				currentTime += time.Duration(e.Len) * time.Second
			}
			vlc.Close()
			prg.Close()
			prg2.WriteString("</table>\n")
			nextDay()
			pls = []*entry{}
		}
	}
	ids := []string{}
	for id, _ := range dates {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		d := dates[id]
		prg2.WriteString(fmt.Sprintf("<a href=\"#%s\">%s</a><br />\n", id, d))
	}
	prg2.WriteString("</body></html>\n")
	prg2.Close()

	logger.Infof("%v items, %v videos, %v hours", items, len(list), seconds/3600)

	/*
		files, err := os.ReadDir(*basedir)
		if err != nil {
			panic(err)
		}
		var fileStrings = []string{}
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), ".mp4") {
				continue
			}
			fileStrings = append(fileStrings, f.Name())
		}
		sort.Sort(sort.Reverse(sort.StringSlice(fileStrings)))
		p, err := os.Create(fmt.Sprintf("%s/playlist.txt", *basedir))
		if err != nil {
			panic(err)
		}
		defer p.Close()

		for key, fs := range fileStrings {
			str := fmt.Sprintf("[Content%v]\n", key)
			str += fmt.Sprintf("File=%s\n", fs)
			str += fmt.Sprintf("Volume=7\n")
			next := key + 1
			if key == len(fileStrings)-1 {
				next = 0
			}
			str += fmt.Sprintf("Succ=%v\n", next)
			p.Write([]byte(str))
		}
	*/
}
