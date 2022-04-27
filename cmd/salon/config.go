package main

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
)

type SalonConfig struct {
	TemplateDev    bool   `toml:"templatedev"`
	BleveIndex     string `toml:"bleveindex"`
	TemplateDir    string `toml:"templatedir"`
	StaticDir      string `toml:"staticdir"`
	PictureFSImage string `toml:"picturefsimage"`
	PictureFSJSON  string `toml:"picturefsjson"`
	ExportPath     string `toml:"exportpath"`
}

type SalonDigitalConfig struct {
	CertPem   string      `toml:"certpem"`
	KeyPem    string      `toml:"keypem"`
	LogFile   string      `toml:"logfile"`
	LogLevel  string      `toml:"loglevel"`
	LogFormat string      `toml:"logformat"`
	AccessLog string      `toml:"accesslog"`
	BaseDir   string      `toml:"basedir"`
	Addr      string      `toml:"addr"`
	AddrExt   string      `toml:"addrext"`
	User      string      `toml:"user"`
	Password  string      `toml:"password"`
	Salon     SalonConfig `toml:"salon"`
}

func LoadSalonDigitalConfig(fp string, conf *SalonDigitalConfig) error {
	_, err := toml.DecodeFile(fp, conf)
	if err != nil {
		return errors.Wrapf(err, "error loading config file %v", fp)
	}
	conf.BaseDir = strings.TrimRight(filepath.ToSlash(conf.BaseDir), "/")
	return nil
}
