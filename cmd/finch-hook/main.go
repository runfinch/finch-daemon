// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/containerd/nerdctl/v2/cmd/nerdctl/helpers"
	"github.com/containerd/nerdctl/v2/pkg/clientutil"
	"github.com/containerd/nerdctl/v2/pkg/logging"
	"github.com/containerd/nerdctl/v2/pkg/ocihook"
	"github.com/spf13/cobra"
)

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
	globalOptions, err := helpers.ProcessRootCmdFlags(cmd)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("event type needs to be passed")
	}
	dataStore, err := clientutil.DataStore(globalOptions.DataRoot, globalOptions.Address)
	if err != nil {
		return err
	}
	return ocihook.Run(os.Stdin, os.Stderr, args[0], dataStore, globalOptions.CNIPath, globalOptions.CNINetConfPath, globalOptions.BridgeIP)
}

func addPersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("debug", false, "debug mode")
	cmd.PersistentFlags().Bool("debug-full", false, "debug mode (with full output)")
	cmd.PersistentFlags().String("address", "", "containerd address")
	cmd.PersistentFlags().String("namespace", "", "containerd namespace")
	cmd.PersistentFlags().String("snapshotter", "", "containerd snapshotter")
	cmd.PersistentFlags().String("cni-path", "", "cni plugins binary directory")
	cmd.PersistentFlags().String("cni-netconfpath", "", "cni config directory")
	cmd.PersistentFlags().String("data-root", "", "Root directory of persistent nerdctl state")
	cmd.PersistentFlags().String("cgroup-manager", "", "Cgroup manager to use")
	cmd.PersistentFlags().Bool("insecure-registry", false, "skips verifying HTTPS certs")
	cmd.PersistentFlags().StringSlice("hosts-dir", nil, "hosts directory")
	cmd.PersistentFlags().Bool("experimental", false, "experimental features")
	cmd.PersistentFlags().String("host-gateway-ip", "", "host gateway IP")
	cmd.PersistentFlags().String("bridge-ip", "", "bridge IP")
	cmd.PersistentFlags().Bool("kube-hide-dupe", false, "deduplicate images for k8s")
	cmd.PersistentFlags().StringSlice("cdi-spec-dirs", nil, "CDI spec directories")
	cmd.PersistentFlags().StringSlice("global-dns", nil, "global DNS servers")
	cmd.PersistentFlags().StringSlice("global-dns-opts", nil, "global DNS options")
	cmd.PersistentFlags().StringSlice("global-dns-search", nil, "global DNS search domains")

}
