package main

import (
	"github.com/BurntSushi/toml"
	"github.com/goph/emperror"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

type SalonConfig struct {
	TemplateDev    bool   `toml:"templatedev"`
	TemplateDir    string `toml:"templatedir"`
	StaticDir      string `toml:"staticdir"`
	PictureFSImage string `toml:"picturefsimage"`
	PictureFSJSON  string `toml:"picturefsjson"`
}

type BangConfig struct {
	TemplateDev bool
	TemplateDir string
}

type SalonDigitalConfig struct {
	CertPem        string      `toml:"certpem"`
	KeyPem         string      `toml:"keypem"`
	LogFile        string      `toml:"logfile"`
	LogLevel       string      `toml:"loglevel"`
	LogFormat      string      `toml:"logformat"`
	AccessLog      string      `toml:"accesslog"`
	BaseDir        string      `toml:"basedir"`
	DataDir        string      `toml:"datadir"`
	BleveIndex     string      `toml:"bleveindex"`
	Addr           string      `toml:"addr"`
	AddrExt        string      `toml:"addrext"`
	User           string      `toml:"user"`
	Password       string      `toml:"password"`
	Salon          SalonConfig `toml:"salon"`
	Browser        bool        `toml:"browser"`
	BrowserTimeout duration    `toml:"browsertimeout"`
	BrowserURL     string      `toml:"browserurl"`
	Bang           BangConfig  `toml:"bang"`
	Station        bool
}

func LoadSalonDigitalConfig(fp string, conf *SalonDigitalConfig) error {
	if _, err := os.Stat(fp); err != nil {
		fp = "/etc/salon-digital.toml"
	}
	_, err := toml.DecodeFile(fp, conf)
	if err != nil {
		return emperror.Wrapf(err, "error loading config file %v", fp)
	}
	conf.BaseDir = strings.TrimRight(filepath.ToSlash(conf.BaseDir), "/")
	return nil
}
