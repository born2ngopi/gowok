package gowok

import (
	"fmt"
	"io"
	"os"

	"github.com/gowok/gowok/config"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App      config.App
	Security config.Security
	SQLs     map[string]config.SQL     `yaml:"sql"`
	MongoDBs map[string]config.MongoDB `yaml:"mongodb"`
	Caches   map[string]config.Cache   `yaml:"cache"`
	Http     map[string]config.Http    `yaml:"http"`
	Smtp     map[string]config.Smtp    `yaml:"smtp"`
	Others   map[string]string         `yaml:"others"`

	IsTesting bool   `yaml:"is_testing"`
	Env       string `yaml:"env"`
}

func NewConfig(pathConfig string) (*Config, error) {
	fiConfig, err := os.OpenFile(pathConfig, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	fiContent, err := io.ReadAll(fiConfig)
	if err != nil {
		return nil, fmt.Errorf("can't read config file: %w", err)
	}

	conf := &Config{}
	err = yaml.Unmarshal(fiContent, conf)
	if err != nil {
		return conf, fmt.Errorf("can't decode config file: %w", err)
	}

	return conf, nil
}
