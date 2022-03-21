package main

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
)

type SalonDigitalConfig struct {
	CertPem        string `toml:"certpem"`
	KeyPem         string `toml:"keypem"`
	LogFile        string `toml:"logfile"`
	LogLevel       string `toml:"loglevel"`
	LogFormat      string `toml:"logformat"`
	AccessLog      string `toml:"accesslog"`
	BaseDir        string `toml:"basedir"`
	StaticDir      string `toml:"staticdir"`
	TemplateDir    string `toml:"templatedir"`
	PictureDir     string `toml:"picturedir"`
	Addr           string `toml:"addr"`
	AddrExt        string `toml:"addrext"`
	User           string `toml:"user"`
	Password       string `toml:"password"`
	ImageTemplate  string `toml:"imagetemplate"`
	PictureFSImage string `toml:"picturefsimage"`
	PictureFSJSON  string `toml:"picturefsjson"`
	TemplateDev    bool   `toml:"templatedev"`
}

func LoadSalonDigitalConfig(fp string, conf *SalonDigitalConfig) error {
	_, err := toml.DecodeFile(fp, conf)
	if err != nil {
		return errors.Wrapf(err, "error loading config file %v", fp)
	}
	conf.BaseDir = strings.TrimRight(filepath.ToSlash(conf.BaseDir), "/")
	return nil
}
