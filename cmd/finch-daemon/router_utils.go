// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/gofrs/flock"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/runfinch/finch-daemon/api/router"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/internal/service/builder"
	"github.com/runfinch/finch-daemon/internal/service/container"
	"github.com/runfinch/finch-daemon/internal/service/distribution"
	"github.com/runfinch/finch-daemon/internal/service/exec"
	"github.com/runfinch/finch-daemon/internal/service/image"
	"github.com/runfinch/finch-daemon/internal/service/network"
	"github.com/runfinch/finch-daemon/internal/service/system"
	"github.com/runfinch/finch-daemon/internal/service/volume"
	"github.com/runfinch/finch-daemon/pkg/archive"
	"github.com/runfinch/finch-daemon/pkg/ecc"
	"github.com/runfinch/finch-daemon/pkg/flog"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// handleConfigOptions gets nerdctl config value from nerdctl.toml file.
func handleConfigOptions(cfg *config.Config, options *DaemonOptions) error {
	tomlPath := options.configPath
	r, err := os.Open(tomlPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // File not found; this is not an error.
		}
		return err // Return other errors directly.
	}
	defer r.Close()

	dec := toml.NewDecoder(r).DisallowUnknownFields()
	if err := dec.Decode(cfg); err != nil {
		return fmt.Errorf(
			"failed to load config from %q : %w",
			tomlPath, err,
		)
	}
	return nil
}

// initializeConfig initializes configuration from file, environment, and set default values.
func initializeConfig(options *DaemonOptions) (*config.Config, error) {
	conf := config.New()

	if err := handleConfigOptions(conf, options); err != nil {
		return nil, err
	}

	if options.debug {
		conf.Debug = options.debug
	}
	if conf.Namespace == "" || conf.Namespace == namespaces.Default {
		conf.Namespace = defaultNamespace
	}

	return conf, nil
}

// createNerdctlWrapper creates the Nerdctl wrapper and checks for the nerdctl binary.
func createNerdctlWrapper(clientWrapper *backend.ContainerdClientWrapper, conf *config.Config) (*backend.NerdctlWrapper, error) {
	// GlobalCommandOptions is actually just an alias for Config, see
	// https://github.com/containerd/nerdctl/blob/9f8655f7722d6e6851755123730436bf1a6c9995/pkg/api/types/global.go#L21
	globalOptions := (*types.GlobalCommandOptions)(conf)
	ncWrapper := backend.NewNerdctlWrapper(clientWrapper, globalOptions)
	if _, err := ncWrapper.GetNerdctlExe(); err != nil {
		return nil, fmt.Errorf("failed to find nerdctl binary: %w", err)
	}
	return ncWrapper, nil
}

// createContainerdClient creates and wraps the containerd client.
func createContainerdClient(conf *config.Config) (*backend.ContainerdClientWrapper, error) {
	client, err := containerd.New(conf.Address, containerd.WithDefaultNamespace(conf.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}
	return backend.NewContainerdClientWrapper(client), nil
}

// sanitizeRegoFile validates and prepares the Rego policy file for use.
// It checks validates the file, acquires a file lock,
// and sets rego file to be read-only.
func sanitizeRegoFile(options *DaemonOptions) (string, error) {
	if options.regoFilePath != "" {
		if !options.enableMiddleware {
			return "", fmt.Errorf("rego file path was provided without the --enable-middleware flag, please provide the --enable-middleware flag") // todo, can we default to setting this flag ourselves is this better UX?
		}

		if err := checkRegoFileValidity(options.regoFilePath); err != nil {
			return "", err
		}
	}

	if options.enableMiddleware && options.regoFilePath == "" {
		return "", fmt.Errorf("rego file path not provided, please provide the policy file path using the --rego-file flag")
	}

	fileLock := flock.New(options.regoFilePath)

	locked, err := fileLock.TryLock()
	if err != nil {
		return "", fmt.Errorf("error acquiring lock on rego file: %v", err)
	}
	if !locked {
		return "", fmt.Errorf("unable to acquire lock on rego file, it may be in use by another process")
	}

	// Change file permissions to read-only
	err = os.Chmod(options.regoFilePath, 0400)
	if err != nil {
		fileLock.Unlock()
		return "", fmt.Errorf("error changing rego file permissions: %v", err)
	}
	options.regoFileLock = fileLock

	return options.regoFilePath, nil
}

// createRouterOptions creates router options by initializing all required services.
func createRouterOptions(
	conf *config.Config,
	clientWrapper *backend.ContainerdClientWrapper,
	ncWrapper *backend.NerdctlWrapper,
	logger *flog.Logrus,
	regoFilePath string,
) *router.Options {
	fs := afero.NewOsFs()
	tarCreator := archive.NewTarCreator(ecc.NewExecCmdCreator(), logger)
	tarExtractor := archive.NewTarExtractor(ecc.NewExecCmdCreator(), logger)

	return &router.Options{
		Config:              conf,
		ContainerService:    container.NewService(clientWrapper, ncWrapper, logger, fs, tarCreator, tarExtractor),
		ImageService:        image.NewService(clientWrapper, ncWrapper, logger),
		NetworkService:      network.NewService(clientWrapper, ncWrapper, logger),
		SystemService:       system.NewService(clientWrapper, ncWrapper, logger),
		BuilderService:      builder.NewService(clientWrapper, ncWrapper, logger, tarExtractor),
		VolumeService:       volume.NewService(ncWrapper, logger),
		ExecService:         exec.NewService(clientWrapper, logger),
		DistributionService: distribution.NewService(clientWrapper, ncWrapper, logger),
		NerdctlWrapper:      ncWrapper,
		RegoFilePath:        regoFilePath,
	}
}

// checkRegoFileValidity verifies that the given rego file exists and has the right file extension.
func checkRegoFileValidity(regoFilePath string) error {
	fmt.Println("filepath in checkRegoFileValidity = ", regoFilePath)
	if _, err := os.Stat(regoFilePath); os.IsNotExist(err) {
		return fmt.Errorf("provided Rego file path does not exist: %s", regoFilePath)
	}

	// Check if the file has a valid extension (.rego)
	fileExt := strings.ToLower(filepath.Ext(regoFilePath))

	fmt.Println("fileExt = ", fileExt)
	if fileExt != ".rego" {
		return fmt.Errorf("invalid file extension for Rego file. Only .rego files are supported")
	}

	return nil
}

func cleanupRegoFile(options *DaemonOptions, logger *flog.Logrus) {
	if options.regoFileLock == nil {
		return // Already cleaned up or nothing to clean
	}

	// unlock the rego file
	if err := options.regoFileLock.Unlock(); err != nil {
		logrus.Errorf("failed to unlock Rego file: %v", err)
	}
	logger.Infof("rego file unlocked")

	// make rego file editable
	if err := os.Chmod(options.regoFilePath, 0600); err != nil {
		logrus.Errorf("failed to change file permissions of rego file: %v", err)
	}

	options.regoFileLock = nil
}
