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

// hookOptions holds the only values that finch-hook needs at OCI hook time.
// finch-hook is a single-purpose binary and intentionally does not use
// helpers.ProcessRootCmdFlags from nerdctl — that helper requires all 19 nerdctl
// global flags to be registered and reads many fields irrelevant to the hook path.
// This struct is scoped to exactly what nerdctl's ocihook.Run and
// clientutil.DataStore actually consume.
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
	rootCmd := &cobra.Command{Use: "finch-hook"}
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
	return ocihook.Run(os.Stdin, os.Stderr, args[0], dataStore, opts.cniPath, opts.cniNetconfPath, opts.bridgeIP)
}

func addPersistentFlags(cmd *cobra.Command) {
	// finch-hook only registers the flags it actually uses. It intentionally does not
	// use helpers.ProcessRootCmdFlags, which requires all 19 nerdctl global flags to be
	// present. These 5 flags are the complete surface area of the finch-hook binary.
	cmd.PersistentFlags().String("address", "", "containerd address")
	cmd.PersistentFlags().String("data-root", "", "root directory of persistent finch-daemon state")
	cmd.PersistentFlags().String("cni-path", "", "cni plugins binary directory")
	cmd.PersistentFlags().String("cni-netconfpath", "", "cni config directory")
	cmd.PersistentFlags().String("bridge-ip", "", "bridge IP")
}

// parseHookOptions reads the 5 flags that finch-hook registers and initialises
// the filesystem helper (fs.InitFS), preserving the side effect that
// helpers.ProcessRootCmdFlags performs in the full nerdctl CLI.
func parseHookOptions(cmd *cobra.Command) (hookOptions, error) {
	var opts hookOptions
	var err error

	if opts.address, err = cmd.Flags().GetString("address"); err != nil {
		return hookOptions{}, err
	}
	if opts.dataRoot, err = cmd.Flags().GetString("data-root"); err != nil {
		return hookOptions{}, err
	}
	if opts.cniPath, err = cmd.Flags().GetString("cni-path"); err != nil {
		return hookOptions{}, err
	}
	if opts.cniNetconfPath, err = cmd.Flags().GetString("cni-netconfpath"); err != nil {
		return hookOptions{}, err
	}
	if opts.bridgeIP, err = cmd.Flags().GetString("bridge-ip"); err != nil {
		return hookOptions{}, err
	}
	if err = fs.InitFS(opts.dataRoot); err != nil {
		return hookOptions{}, err
	}
	return opts, nil
}
