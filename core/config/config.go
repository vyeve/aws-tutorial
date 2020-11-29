package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	LogLevel                      string `yaml:"logLevel"`
	Region                        string `yaml:"region"`
	Profile                       string `yaml:"profile"`
	CredentialsChainVerboseErrors bool   `yaml:"credentialsChainVerboseErrors"`
}

const ConfigPathEnv = "APP_CONFIG"

func New() (*Configuration, error) {
	c := new(Configuration)
	path := filepath.Join(os.Getenv(ConfigPathEnv), "config.yaml")
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close() // nolint: errcheck

	dec := yaml.NewDecoder(f)
	err = dec.Decode(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
