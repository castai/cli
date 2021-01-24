package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

const (
	DefaultHostname   = "api.cast.ai"
	defaultConfigDir  = ".cast"
	defaultConfigName = "config"

	envApiToken    = "CASTAI_API_TOKEN"
	envApiHostname = "CASTAI_API_HOSTNAME"
	envDebug       = "CASTAI_DEBUG"
	envConfigPath  = "CASTAI_CONFIG"
)

type Config struct {
	Hostname    string `yaml:"hostname"`
	AccessToken string `yaml:"access_token"`
	Debug       bool   `yaml:"debug"`
}

func LoadFromEnv() (*Config, error) {
	config := &Config{
		Hostname:    "api.cast.ai",
		AccessToken: "",
		Debug:       false,
	}

	// Try to read config from file.
	configPath, err := GetPath()
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(configPath)
	if err == nil {
		bytes, err := ioutil.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(bytes, &config); err != nil {
			return nil, fmt.Errorf("parsing config file from %q: %w", configPath, err)
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Override with env variables if any.
	if hostname := os.Getenv(envApiHostname); hostname != "" {
		config.Hostname = hostname
	}
	if accessToken := os.Getenv(envApiToken); accessToken != "" {
		config.AccessToken = accessToken
	}
	if debug := os.Getenv(envDebug); debug != "" {
		config.Debug = debug == "true" || debug == "1"
	}

	return config, nil
}

func Save(cfg *Config) error {
	configPath, err := GetPath()
	if err != nil {
		return err
	}
	if err := ensureDir(path.Dir(configPath)); err != nil {
		return fmt.Errorf("ensuring directory exist: %w", err)
	}
	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(configPath, bytes, 0600); err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}
	return nil
}

func ensureDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}
		return nil
	}
	return err
}

func GetPath() (string, error) {
	if p := os.Getenv(envConfigPath); p != "" {
		return p, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(homeDir, defaultConfigDir, defaultConfigName), nil
}
