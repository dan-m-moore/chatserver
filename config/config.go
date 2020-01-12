// Package config provides a utility to parse a JSON config file into
// a struct containing the parsed data.
package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

// Config contains configuration data.
type Config struct {
	TelnetPort    int
	WebPort       int
	WebClientPath string
	LogFilePath   string
}

// ParseFile attempts to open a JSON config file at a given location, parse it
// into a Config struct, validate the contents, and return the data.
func ParseFile(configFilePath string) (*Config, error) {
	// Read the config file
	configData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	// Parse the config JSON
	config := Config{}
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, errors.New("invalid config file")
	}

	// Validate the ports
	if config.TelnetPort <= 0 {
		return nil, errors.New("invalid telnet port")
	}

	if config.WebPort <= 0 {
		return nil, errors.New("invalid web port")
	}

	// Validate the web client path
	info, err := os.Stat(config.WebClientPath)
	if (err != nil && os.IsNotExist(err)) || !info.IsDir() {
		return nil, errors.New("invalid web client path")
	}

	return &config, nil
}
