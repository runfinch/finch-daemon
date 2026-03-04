// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/containerd/nerdctl/v2/pkg/clientutil"
	"github.com/containerd/nerdctl/v2/pkg/fs"
	"github.com/containerd/nerdctl/v2/pkg/logging"
	"github.com/containerd/nerdctl/v2/pkg/ocihook"
	"github.com/spf13/cobra"
)

// hookOptions holds the values consumed by the OCI lifecycle hook path
// (internal oci-hook). finch-shim does not use helpers.ProcessRootCmdFlags
// from nerdctl because that helper requires all 19 global flags, most of
// which are irrelevant to hook execution. These 5 fields are scoped to
// exactly what ocihook.Run and clientutil.DataStore consume.
type hookOptions struct {
	address        string
	dataRoot       string
	cniPath        string
	cniNetconfPath string
	bridgeIP       string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) == 3 && os.Args[1] == logging.MagicArgv1 {
		return logging.Main(os.Args[2])
	}
	return newApp().Execute()
}

func newApp() *cobra.Command {
	rootCmd := &cobra.Command{Use: "finch-shim"}
	addPersistentFlags(rootCmd)

	internalCmd := &cobra.Command{Use: "internal"}
	ociHookCmd := &cobra.Command{
		Use:           "oci-hook",
		Short:         "OCI hook",
		RunE:          internalOCIHookAction,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	internalCmd.AddCommand(ociHookCmd)
	rootCmd.AddCommand(internalCmd)
	return rootCmd
}

func internalOCIHookAction(cmd *cobra.Command, args []string) error {
	opts, err := parseHookOptions(cmd)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("event type needs to be passed")
	}
	dataStore, err := clientutil.DataStore(opts.dataRoot, opts.address)
	if err != nil {
		return err
	}
	return ocihook.Run(os.Stdin, os.Stderr, args[0], dataStore,
		opts.cniPath, opts.cniNetconfPath, opts.bridgeIP)
}

// addPersistentFlags adds only the flags consumed by the OCI hook path.
// See the hookOptions struct.
func addPersistentFlags(cmd *cobra.Command) {
	//
	cmd.PersistentFlags().String("address", "", "containerd address")
	cmd.PersistentFlags().String("data-root", "", "root directory of persistent finch-daemon state")
	cmd.PersistentFlags().String("cni-path", "", "cni plugins binary directory")
	cmd.PersistentFlags().String("cni-netconfpath", "", "cni config directory")
	cmd.PersistentFlags().String("bridge-ip", "", "bridge IP")
}

// parseHookOptions reads the hook flags and calls fs.InitFS to preserve
// the side effect that helpers.ProcessRootCmdFlags performs in nerdctl.
func parseHookOptions(cmd *cobra.Command) (hookOptions, error) {
	var opts hookOptions
	var err error
	get := func(name string, dest *string) {
		if err == nil {
			*dest, err = cmd.Flags().GetString(name)
		}
	}
	get("address", &opts.address)
	get("data-root", &opts.dataRoot)
	get("cni-path", &opts.cniPath)
	get("cni-netconfpath", &opts.cniNetconfPath)
	get("bridge-ip", &opts.bridgeIP)
	if err != nil {
		return hookOptions{}, err
	}
	if err = fs.InitFS(opts.dataRoot); err != nil {
		return hookOptions{}, err
	}
	return opts, nil
}
