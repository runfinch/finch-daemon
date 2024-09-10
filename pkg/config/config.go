// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type Config struct {
	RegoFilePath string `yaml:"rego_file_path,omitempty"`
}

// Load reads a YAML file from a given location and returns a new Config struct.
func Load(cfgPath string, fs afero.Fs) (*Config, error) {
	b, err := afero.ReadFile(fs, cfgPath)
	if err != nil {
		// Ignore file not found errors
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return &Config{}, err
	}

	cfg := CreateDefaultConfig()
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return &Config{}, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return cfg, nil
}

func CreateDefaultConfig() *Config {
	return &Config{
		RegoFilePath: "",
	}
}
