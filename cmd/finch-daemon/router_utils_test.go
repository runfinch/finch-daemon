// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeConfig(t *testing.T) {
	options := &DaemonOptions{}
	options.debug = true

	cfg, err := initializeConfig(options)
	require.NoError(t, err, "Initialization should succeed.")

	assert.True(t, cfg.Debug, "Debug mode should be enabled.")
	assert.Equal(t, "finch", defaultNamespace, "check default namespace")
}

func TestHandleConfigOptions_FileNotFound(t *testing.T) {
	cfg := &config.Config{}
	options := &DaemonOptions{}
	options.configPath = "/non/existing/path/nerdctl.toml"

	err := handleConfigOptions(cfg, options)
	assert.NoError(t, err, "File not found should not cause an error.")
}

func TestHandleConfigOptions_InvalidTOML(t *testing.T) {
	cfg := &config.Config{}
	options := &DaemonOptions{}

	tmpFile, err := os.CreateTemp("/tmp", "invalid.toml")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	options.configPath = tmpFile.Name()

	_, _ = tmpFile.WriteString("invalid_toml")

	err = handleConfigOptions(cfg, options)
	assert.Error(t, err, "Invalid TOML should cause an error.")
}

func TestHandleConfigOptions_ValidTOML(t *testing.T) {
	cfg := &config.Config{}
	options := &DaemonOptions{}

	// Create a temporary valid TOML file
	tmpFile, err := os.CreateTemp("", "valid.toml")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	options.configPath = tmpFile.Name()

	_, _ = tmpFile.WriteString(`
address = "test_address"
namespace = "test_namespace"
`)

	err = handleConfigOptions(cfg, options)
	assert.NoError(t, err, "Valid TOML should not cause an error.")
	assert.Equal(t, "test_address", cfg.Address)
	assert.Equal(t, "test_namespace", cfg.Namespace)
}
