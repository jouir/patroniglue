package main

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config stores configuration
type Config struct {
	File     string
	Frontend FrontendConfig `yaml:"frontend"`
	Backend  BackendConfig  `yaml:"backend"`
	Cache    CacheConfig    `yaml:"cache"`
}

// FrontendConfig for storing Frontend settings
type FrontendConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Certfile  string `yaml:"certfile"`
	Keyfile   string `yaml:"keyfile"`
	LogFormat string `yaml:"logformat"`
}

// BackendConfig for storing Backend settings
type BackendConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Scheme   string `yaml:"scheme"`
	Insecure bool   `yaml:"insecure"`
}

// CacheConfig for storing Cache settings
type CacheConfig struct {
	TTL      float64 `yaml:"ttl"`
	Interval float64 `yaml:"interval"`
}

// NewConfig creates a Config
func NewConfig() *Config {
	return &Config{}
}

// ReadFile reads a configuration file and load settings to memory
func (c *Config) ReadFile(file string) error {
	file, err := filepath.Abs(file)
	if err != nil {
		return err
	}

	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return err
	}

	return nil
}
