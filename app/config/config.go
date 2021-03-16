package config

import (
	"cloud.google.com/go/compute/metadata"
	"golang.org/x/xerrors"
)

var gConf *Config

func init() {
	gConf = defaultConfig()
}

type Config struct {
	ProjectID   string
	Port        int
	DevelopMode bool
}

func defaultConfig() *Config {
	var conf Config
	conf.DevelopMode = !metadata.OnGCE()
	conf.Port = 8080
	conf.ProjectID = "fes"
	return &conf
}

func Set(opts []Option) error {
	for idx, opt := range opts {
		err := opt(gConf)
		if err != nil {
			return xerrors.Errorf("option setting[%d:%T] error: %w", idx, opt, err)
		}
	}
	return nil
}

func Get() *Config {
	return gConf
}
