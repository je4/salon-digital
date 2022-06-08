package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	var basedir = flag.String("basedir", ".", "base folder with html contents")

	flag.Parse()

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

}
