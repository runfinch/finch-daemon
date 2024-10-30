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
	"strings"
	"sync"
	"syscall"
	"time"

	// #nosec
	// register HTTP handler for /debug/pprof on the DefaultServeMux.
	_ "net/http/pprof"

	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/runfinch/finch-daemon/api/router"
	"github.com/runfinch/finch-daemon/pkg/flog"
	"github.com/runfinch/finch-daemon/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// Keep this value in sync with `guestSocket` in README.md.
	defaultFinchAddr  = "/run/finch.sock"
	defaultNamespace  = "finch"
	defaultConfigPath = "/etc/finch/finch.toml"
)

type DaemonOptions struct {
	debug        bool
	socketAddr   string
	socketOwner  int
	debugAddress string
	configPath   string
}

var options = new(DaemonOptions)

func main() {
	rootCmd := &cobra.Command{
		Use:          "finch-daemon",
		Short:        "Finch daemon with a Docker-compatible API",
		Version:      strings.TrimPrefix(version.Version, "v"),
		RunE:         runAdapter,
		SilenceUsage: true,
	}
	rootCmd.Flags().StringVar(&options.socketAddr, "socket-addr", defaultFinchAddr, "server listening Unix socket address")
	rootCmd.Flags().BoolVar(&options.debug, "debug", false, "turn on debug log level")
	rootCmd.Flags().IntVar(&options.socketOwner, "socket-owner", -1, "Uid and Gid of the server socket")
	rootCmd.Flags().StringVar(&options.debugAddress, "debug-addr", "", "")
	rootCmd.Flags().StringVar(&options.configPath, "config-file", defaultConfigPath, "Daemon Config Path")
	if err := rootCmd.Execute(); err != nil {
		log.Printf("got error: %v", err)
		log.Fatal(err)
	}
}

func runAdapter(cmd *cobra.Command, _ []string) error {
	return run(options)
}

func run(options *DaemonOptions) error {
	// This sets the log level of the dependencies that use logrus (e.g., containerd library).
	if options.debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logger := flog.NewLogrus()
	r, err := newRouter(options, logger)
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

	if options.debugAddress != "" {
		logger.Infof("Serving debugging endpoint on %q", options.debugAddress)
		go func() {
			debugListener, err := net.Listen("tcp", options.debugAddress)
			if err != nil {
				logger.Fatal(err)
			}
			debugServer := &http.Server{
				Handler:           http.DefaultServeMux,
				ReadHeaderTimeout: 5 * time.Second,
			}
			if err := debugServer.Serve(debugListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Fatal(err)
			}
		}()
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

func newRouter(options *DaemonOptions, logger *flog.Logrus) (http.Handler, error) {
	conf, err := initializeConfig(options)
	if err != nil {
		return nil, err
	}

	clientWrapper, err := createContainerdClient(conf)
	if err != nil {
		return nil, err
	}

	ncWrapper, err := createNerdctlWrapper(clientWrapper, conf)
	if err != nil {
		return nil, err
	}

	opts := createRouterOptions(conf, clientWrapper, ncWrapper, logger)
	return router.New(opts), nil
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
