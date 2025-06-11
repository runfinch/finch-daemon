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
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	// #nosec
	// register HTTP handler for /debug/pprof on the DefaultServeMux.
	_ "net/http/pprof"

	"github.com/coreos/go-systemd/v22/activation"
	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/gofrs/flock"
	"github.com/moby/moby/pkg/pidfile"
	"github.com/runfinch/finch-daemon/api/router"
	"github.com/runfinch/finch-daemon/internal/fs/passwd"
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
	defaultPidFile    = "/run/finch.pid"
)

type DaemonOptions struct {
	debug              bool
	socketAddr         string
	socketOwner        int
	debugAddress       string
	configPath         string
	pidFile            string
	regoFilePath       string
	enableExperimental bool
	skipRegoPermCheck  bool
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
	rootCmd.Flags().StringVar(&options.pidFile, "pidfile", defaultPidFile, "pid file location")
	rootCmd.Flags().StringVar(&options.regoFilePath, "rego-file", "", "Rego Policy Path (requires --experimental flag)")
	rootCmd.Flags().BoolVar(&options.skipRegoPermCheck, "skip-rego-perm-check", false, "skip the rego file permission check (allows permissions more permissive than 0600)")
	rootCmd.Flags().BoolVar(&options.enableExperimental, "experimental", false, "enable experimental features")

	if err := rootCmd.Execute(); err != nil {
		log.Printf("got error: %v", err)
		log.Fatal(err)
	}
}

func runAdapter(cmd *cobra.Command, _ []string) error {
	return run(options)
}

func getListener(options *DaemonOptions) (net.Listener, error) {
	var listener net.Listener
	var err error

	if options.socketAddr == "fd://" {
		if options.socketOwner != -1 {
			return nil, fmt.Errorf("socket-owner is not supported while using socket activation using fd://")
		}

		listeners, err := activation.Listeners()
		if err != nil {
			return nil, fmt.Errorf("cannot retrieve listeners: %w", err)
		}
		if len(listeners) != 1 {
			return nil, fmt.Errorf("unexpected number of socket activations (%d != 1)", len(listeners))
		}
		listener = listeners[0]
	} else {
		listener, err = net.Listen("unix", options.socketAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to listen on %s: %w", options.socketAddr, err)
		}
		if err := os.Chown(options.socketAddr, options.socketOwner, options.socketOwner); err != nil {
			return nil, fmt.Errorf("failed to chown the socket: %w", err)
		}
	}
	return listener, nil
}

func run(options *DaemonOptions) error {
	// This sets the log level of the dependencies that use logrus (e.g., containerd library).
	if options.debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if options.pidFile != "" {
		if err := os.MkdirAll(filepath.Dir(options.pidFile), 0o600); err != nil {
			return fmt.Errorf("failed to create pidfile directory %s", err)
		}

		pidFileLock := flock.New(options.pidFile)

		defer func() {
			pidFileLock.Unlock()
		}()

		if isLocked, err := pidFileLock.TryLock(); err != nil || !isLocked {
			return fmt.Errorf("failed to acquire lock on PID file (%s); ensure only one instance is using the PID file path", options.pidFile)
		}

		if err := pidfile.Write(options.pidFile, os.Getpid()); err != nil {
			return fmt.Errorf("failed to start daemon, ensure finch daemon is not running or delete %s %w", options.pidFile, err)
		}

		pidFileLock.Unlock()

		// Defer is at the end of trying to write to PID file so that it doesn't remove the pidfile created by another process when daemon fails to start
		defer func() {
			if err := os.Remove(options.pidFile); err != nil {
				logrus.Errorf("failed to remove pidfile %s", options.pidFile)
			}
		}()
	}

	logger := flog.NewLogrus()
	r, err := newRouter(options, logger)
	if err != nil {
		return fmt.Errorf("failed to create a router: %w", err)
	}

	serverWg := &sync.WaitGroup{}
	serverWg.Add(1)

	if options.socketOwner >= 0 {
		if err := defineDockerConfig(options.socketOwner); err != nil {
			return fmt.Errorf("failed to get finch config: %w", err)
		}
	}

	listener, err := getListener(options)
	if err != nil {
		return fmt.Errorf("failed to create a listener: %w", err)
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

	var regoFilePath string

	if options.regoFilePath != "" {
		if !options.enableExperimental {
			return nil, fmt.Errorf("rego file provided without experimental flag - OPA middleware is an experimental feature, please enable it with '--experimental' flag")
		}
		regoFilePath, err = checkRegoFileValidity(options, logger)
		if err != nil {
			return nil, err
		}
	} else if options.enableExperimental {
		// Only experimental flag set
		logger.Info("experimental flag passed, but no experimental features enabled")
	}

	opts := createRouterOptions(conf, clientWrapper, ncWrapper, logger, regoFilePath)
	newRouter, err := router.New(opts)
	if err != nil {
		return nil, err
	}
	return newRouter, nil
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

// defineDockerConfig defines the DOCKER_CONFIG environment variable for the process
// to be $HOME/.finch. When building an image via finch-daemon, buildctl uses this variable
// to load auth configs.
//
// This is a hack and should be fixed by passing the actual credentials that come in
// via the build API to buildctl instead.
func defineDockerConfig(uid int) error {
	return passwd.Walk(func(e passwd.Entry) bool {
		if e.UID == uid {
			os.Setenv("DOCKER_CONFIG", fmt.Sprintf("%s/.finch", e.Home))
			return false
		}
		return true
	})
}
