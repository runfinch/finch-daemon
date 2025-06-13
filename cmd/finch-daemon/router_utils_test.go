// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/runfinch/finch-daemon/pkg/flog"
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

func TestCheckRegoFileValidity(t *testing.T) {
	logger := flog.NewLogrus()
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) (string, func())
		skipPermCheck bool
		expectedError string
	}{
		{
			name: "valid rego file",
			setupFunc: func(t *testing.T) (string, func()) {
				// Create a temporary directory that will be automatically cleaned up
				tmpDir := t.TempDir()

				// Create a file with .rego extension and proper content
				regoPath := filepath.Join(tmpDir, "test.rego")
				regoContent := `package finch.authz

import future.keywords.if
import rego.v1

default allow = false
`
				err := os.WriteFile(regoPath, []byte(regoContent), 0600)
				require.NoError(t, err)

				return regoPath, func() {}
			},
			expectedError: "",
		},
		{
			name: "non-existent file",
			setupFunc: func(t *testing.T) (string, func()) {
				return filepath.Join(os.TempDir(), "nonexistent.rego"), func() {}
			},
			expectedError: "provided Rego file path does not exist",
		},
		{
			name: "wrong extension",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpFile, err := os.CreateTemp("", "test.txt")
				require.NoError(t, err)
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectedError: "invalid file extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, cleanup := tt.setupFunc(t)
			defer cleanup()

			options := &DaemonOptions{
				regoFilePath: filePath,
			}
			path, err := checkRegoFileValidity(options, logger)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Empty(t, path)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, filePath, path)
			}
		})
	}
}
