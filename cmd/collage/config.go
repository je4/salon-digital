package main

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
)

// main config structure for toml file
type CollageConfig struct {
	BaseDir   string            `toml:"basedir"`
	OutputDir string            `toml:"outputdir"`
	Image     map[string]string `toml:"image"`
}

func LoadCollageConfig(fp string, conf *CollageConfig) error {
	_, err := toml.DecodeFile(fp, conf)
	if err != nil {
		return errors.Wrapf(err, "error loading config file %v", fp)
	}
	conf.BaseDir = strings.TrimRight(filepath.ToSlash(conf.BaseDir), "/")
	return nil
}
