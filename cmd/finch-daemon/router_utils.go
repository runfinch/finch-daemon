// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/nerdctl/pkg/api/types"
	"github.com/containerd/nerdctl/pkg/config"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/runfinch/finch-daemon/api/router"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/internal/service/builder"
	"github.com/runfinch/finch-daemon/internal/service/container"
	"github.com/runfinch/finch-daemon/internal/service/exec"
	"github.com/runfinch/finch-daemon/internal/service/image"
	"github.com/runfinch/finch-daemon/internal/service/network"
	"github.com/runfinch/finch-daemon/internal/service/system"
	"github.com/runfinch/finch-daemon/internal/service/volume"
	"github.com/runfinch/finch-daemon/pkg/archive"
	"github.com/runfinch/finch-daemon/pkg/ecc"
	"github.com/runfinch/finch-daemon/pkg/flog"
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

// createRouterOptions creates router options by initializing all required services.
func createRouterOptions(
	conf *config.Config,
	clientWrapper *backend.ContainerdClientWrapper,
	ncWrapper *backend.NerdctlWrapper,
	logger *flog.Logrus,
) *router.Options {
	fs := afero.NewOsFs()
	tarCreator := archive.NewTarCreator(ecc.NewExecCmdCreator(), logger)
	tarExtractor := archive.NewTarExtractor(ecc.NewExecCmdCreator(), logger)

	return &router.Options{
		Config:           conf,
		ContainerService: container.NewService(clientWrapper, ncWrapper, logger, fs, tarCreator, tarExtractor),
		ImageService:     image.NewService(clientWrapper, ncWrapper, logger),
		NetworkService:   network.NewService(clientWrapper, ncWrapper, logger),
		SystemService:    system.NewService(clientWrapper, ncWrapper, logger),
		BuilderService:   builder.NewService(clientWrapper, ncWrapper, logger, tarExtractor),
		VolumeService:    volume.NewService(ncWrapper, logger),
		ExecService:      exec.NewService(clientWrapper, logger),
		NerdctlWrapper:   ncWrapper,
	}
}
