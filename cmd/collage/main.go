package main

import (
	"flag"
	"github.com/pkg/errors"
	"image"
	"log"
	"os"
)

func loadImage(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open image %s", filePath)
	}
	defer f.Close()

	image, _, err := image.Decode(f)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot decode image %s", filePath)
	}

	return image, nil
}

func main() {
	var basedir = flag.String("basedir", ".", "base folder with html contents")
	var configfile = flag.String("cfg", "/etc/tbbs.toml", "configuration file")

	flag.Parse()

	var config = &CollageConfig{
		BaseDir: *basedir,
	}
	if err := LoadCollageConfig(*configfile, config); err != nil {
		log.Fatalf("cannot load config file %s: %v", *configfile, err)
	}

	for name, filePath := range config.Image {
		image, err := loadImage(filePath)
		if err != nil {
			log.Fatal(err)
		}
		bounds := image.Bounds()
	}
}
