// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/nerdctl/pkg/api/types"
	"github.com/containerd/nerdctl/pkg/config"
	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

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
	daemonConfig "github.com/runfinch/finch-daemon/pkg/config"
	"github.com/runfinch/finch-daemon/pkg/ecc"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

const (
	// Keep this value in sync with `guestSocket` in README.md.
	defaultFinchAddr = "/run/finch.sock"
	defaultNamespace = "finch"
)

type DaemonOptions struct {
	debug        bool
	socketAddr   string
	socketOwner  int
	regoFilePath string
	configPath   string
}

var options = new(DaemonOptions)

func main() {
	rootCmd := &cobra.Command{
		Use:          "finch-daemon",
		Short:        "Docker Engine API backed by containerd in finch VM",
		RunE:         runAdapter,
		SilenceUsage: true,
	}
	rootCmd.Flags().StringVar(&options.socketAddr, "socketAddr", defaultFinchAddr, "Server listen socket address. Currently only supports UNIX socket paths.")
	rootCmd.Flags().BoolVar(&options.debug, "debug", false, "whether to print debug logs")
	rootCmd.Flags().IntVar(&options.socketOwner, "socket-owner", -1, "UID and GID of the socket to which finch-daemon will listen."+
		" It's useful when finch-daemon needs to be run as root to access other resources (e.g., rootful containerd socket),"+
		" For macOS, the socket has to be owned by the lima user to make port forwarding work"+
		" (more info: https://github.com/lima-vm/lima/blob/5a9bca3d09481ed7109b14f8d3f0074816731f43/examples/default.yaml#L340)."+
		" -1 means no-op.")
	rootCmd.Flags().StringVar(&options.regoFilePath, "rego-path", "", "Optional path to a rego policy. Currently only allowlist/denylist options are available")
	rootCmd.Flags().StringVar(&options.configPath, "config", "", "Optional path to a settings YAML file (finch-daemon.yaml)")
	if err := rootCmd.Execute(); err != nil {
		log.Printf("got error: %v", err)
		log.Fatal(err)
	}
}

func runAdapter(cmd *cobra.Command, _ []string) error {
	return run(options)
}

func run(options *DaemonOptions) error {
	fs := afero.NewOsFs()

	if options.configPath != "" {
		cfg, err := daemonConfig.Load(options.configPath, fs)
		if err != nil {
			return fmt.Errorf("could not read from %v: %v", options.configPath, err)
		}

		if options.regoFilePath == "" {
			options.regoFilePath = cfg.RegoFilePath
		}
	}

	// This sets the log level of the dependencies that use logrus (e.g., containerd library).
	if options.debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logger := flog.NewLogrus()
	r, err := newRouter(options, logger, fs)
	if err != nil {
		return fmt.Errorf("failed to create a router: %w", err)
	}

	serverWg := &sync.WaitGroup{}
	serverWg.Add(1)

	listener, err := net.Listen("unix", options.socketAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", options.socketAddr, err)
	}
	// TODO: Revisit this after we use systemd to manage finch-daemon.
	// Related: https://github.com/lima-vm/lima/blob/5a9bca3d09481ed7109b14f8d3f0074816731f43/examples/podman-rootful.yaml#L44
	if err := os.Chown(options.socketAddr, options.socketOwner, options.socketOwner); err != nil {
		return fmt.Errorf("failed to chown the finch-daemon socket: %w", err)
	}
	server := &http.Server{
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Minute,
	}
	handleSignal(options.socketAddr, server, logger)

	go func() {
		logger.Infof("Serving on %s...", options.socketAddr)
		defer serverWg.Done()
		// Serve will either exit with an error immediately or return
		// http.ErrServerClosed when the server is successfully closed.
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(err)
		}
	}()

	sdNotify(daemon.SdNotifyReady, logger)
	serverWg.Wait()
	logger.Debugln("Server stopped. Exiting...")
	return nil
}

func newRouter(options *DaemonOptions, logger *flog.Logrus, fs afero.Fs) (http.Handler, error) {
	conf := config.New()
	conf.Debug = options.debug
	conf.Namespace = defaultNamespace
	client, err := containerd.New(conf.Address, containerd.WithDefaultNamespace(conf.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}
	clientWrapper := backend.NewContainerdClientWrapper(client)
	// GlobalCommandOptions is actually just an alias for Config, see
	// https://github.com/containerd/nerdctl/blob/9f8655f7722d6e6851755123730436bf1a6c9995/pkg/api/types/global.go#L21
	globalOptions := (*types.GlobalCommandOptions)(conf)
	ncWrapper := backend.NewNerdctlWrapper(clientWrapper, globalOptions)
	if _, err = ncWrapper.GetNerdctlExe(); err != nil {
		return nil, fmt.Errorf("failed to find nerdctl binary: %w", err)
	}
	execCmdCreator := ecc.NewExecCmdCreator()
	tarCreator := archive.NewTarCreator(execCmdCreator, logger)
	tarExtractor := archive.NewTarExtractor(execCmdCreator, logger)
	opts := &router.Options{
		Config:           conf,
		ContainerService: container.NewService(clientWrapper, ncWrapper, logger, fs, tarCreator, tarExtractor),
		ImageService:     image.NewService(clientWrapper, ncWrapper, logger),
		NetworkService:   network.NewService(clientWrapper, ncWrapper, logger),
		SystemService:    system.NewService(clientWrapper, ncWrapper, logger),
		BuilderService:   builder.NewService(clientWrapper, ncWrapper, logger, tarExtractor),
		VolumeService:    volume.NewService(ncWrapper, logger),
		ExecService:      exec.NewService(clientWrapper, logger),
		RegoFilePath:     options.regoFilePath,
		NerdctlWrapper:   ncWrapper,
	}
	return router.New(opts)
}

func handleSignal(socket string, server *http.Server, logger *flog.Logrus) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		switch sig {
		case os.Interrupt:
			sdNotify(daemon.SdNotifyStopping, logger)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				log.Fatal(err)
			}
		case syscall.SIGTERM:
			sdNotify(daemon.SdNotifyStopping, logger)
			if err := server.Close(); err != nil {
				log.Fatal(err)
			}
			os.Remove(socket)
		}
	}()
}

func sdNotify(state string, logger *flog.Logrus) {
	// (false, nil) - notification not supported (i.e. NOTIFY_SOCKET is unset)
	// (false, err) - notification supported, but failure happened (e.g. error connecting to NOTIFY_SOCKET or while sending data)
	// (true, nil) - notification supported, data has been sent
	notified, err := daemon.SdNotify(false, state)
	logger.Debugf("systemd-notify result: (signaled %t), (err: %v)", notified, err)
}
