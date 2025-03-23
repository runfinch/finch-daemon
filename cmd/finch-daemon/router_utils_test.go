// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/gofrs/flock"
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

func TestCleanupRegoFile(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (*DaemonOptions, *flog.Logrus, func())
	}{
		{
			name: "successful cleanup",
			setupFunc: func() (*DaemonOptions, *flog.Logrus, func()) {
				tmpFile, err := os.CreateTemp("", "test.rego")
				require.NoError(t, err)

				fileLock := flock.New(tmpFile.Name())
				_, err = fileLock.TryLock()
				require.NoError(t, err)

				err = os.Chmod(tmpFile.Name(), 0400)
				require.NoError(t, err)

				logger := flog.NewLogrus()

				cleanup := func() {
					os.Remove(tmpFile.Name())
				}

				return &DaemonOptions{
					regoFilePath: tmpFile.Name(),
					regoFileLock: fileLock,
				}, logger, cleanup
			},
		},
		{
			name: "nil lock handle",
			setupFunc: func() (*DaemonOptions, *flog.Logrus, func()) {
				return &DaemonOptions{
					regoFileLock: nil,
				}, flog.NewLogrus(), func() {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options, logger, cleanup := tt.setupFunc()
			defer cleanup()

			cleanupRegoFile(options, logger)

			if options.regoFilePath != "" {
				// Verify file permissions are restored
				info, err := os.Stat(options.regoFilePath)
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
			}

			// Verify lock is released
			assert.Nil(t, options.regoFileLock)
		})
	}
}

func TestCheckRegoFileValidity(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() (string, func())
		expectedError string
	}{
		{
			name: "valid rego file",
			setupFunc: func() (string, func()) {
				// Create a temporary directory
				tmpDir, err := os.MkdirTemp("", "rego_test")
				require.NoError(t, err)

				// Create a file with .rego extension and proper content
				regoPath := filepath.Join(tmpDir, "test.rego")
				regoContent := `package finch.authz

import future.keywords.if
import rego.v1

default allow = false
`
				fmt.Println("regopath = ", regoPath)
				err = os.WriteFile(regoPath, []byte(regoContent), 0600)
				require.NoError(t, err)

				return regoPath, func() {
					os.RemoveAll(tmpDir)
				}
			},
			expectedError: "",
		},
		{
			name: "non-existent file",
			setupFunc: func() (string, func()) {
				return filepath.Join(os.TempDir(), "nonexistent.rego"), func() {}
			},
			expectedError: "provided Rego file path does not exist",
		},
		{
			name: "wrong extension",
			setupFunc: func() (string, func()) {
				tmpFile, err := os.CreateTemp("", "test.txt")
				require.NoError(t, err)
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectedError: "invalid file extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, cleanup := tt.setupFunc()
			defer cleanup()

			err := checkRegoFileValidity(filePath)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
