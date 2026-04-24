// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp_CommandStructure(t *testing.T) {
	app := newApp()
	assert.Equal(t, "finch-shim", app.Use)

	// Verify "internal" subcommand exists
	internalCmd, _, err := app.Find([]string{"internal"})
	require.NoError(t, err, "'internal' subcommand should exist")
	assert.Equal(t, "internal", internalCmd.Use)

	// Verify "internal oci-hook" subcommand exists
	ociHookCmd, _, err := app.Find([]string{"internal", "oci-hook"})
	require.NoError(t, err, "'internal oci-hook' subcommand should exist")
	assert.Equal(t, "oci-hook", ociHookCmd.Use)
	assert.NotNil(t, ociHookCmd.RunE, "oci-hook command should have RunE set")
}

func TestNewApp_ParsesOCIHookArgs(t *testing.T) {
	app := newApp()
	app.SetArgs([]string{
		"--address=/run/containerd/containerd.sock",
		"--data-root=/tmp/test",
		"--cni-path=/opt/cni/bin",
		"--cni-netconfpath=/etc/cni/net.d",
		"internal", "oci-hook", "createRuntime",
	})

	// Execute will fail at ocihook.Run (no valid OCI state on stdin),
	// but flag parsing and command routing should succeed.
	err := app.Execute()
	require.Error(t, err)
	// The error should NOT be a flag parsing error
	assert.NotContains(t, err.Error(), "unknown flag")
}

func TestNewApp_ParsesOCIHookArgsWithBridgeIP(t *testing.T) {
	app := newApp()
	app.SetArgs([]string{
		"--address=/run/containerd/containerd.sock",
		"--data-root=/tmp/test",
		"--cni-path=/opt/cni/bin",
		"--cni-netconfpath=/etc/cni/net.d",
		"--bridge-ip=10.4.0.0/24",
		"internal", "oci-hook", "createRuntime",
	})

	// Same as TestNewApp_ParsesOCIHookArgs but with the optional --bridge-ip flag.
	err := app.Execute()
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag")
}

func TestNewApp_OCIHookRequiresEventType(t *testing.T) {
	app := newApp()
	app.SetArgs([]string{"internal", "oci-hook"})

	err := app.Execute()
	require.Error(t, err)
	assert.Equal(t, "event type needs to be passed", err.Error())
}

func TestAddPersistentFlags(t *testing.T) {
	app := newApp()
	flags := app.PersistentFlags()

	// finch-shim only registers the 5 flags consumed by the OCI hook path.
	// It intentionally does not register the full nerdctl global flagset —
	// see addPersistentFlags and parseHookOptions in main.go.
	requiredFlags := []string{
		"address", "data-root", "cni-path", "cni-netconfpath", "bridge-ip",
	}
	for _, name := range requiredFlags {
		assert.NotNil(t, flags.Lookup(name), "missing required persistent flag %q", name)
	}

	// Flags that belonged to the old nerdctl global flagset but are not consumed
	// by finch-shim should NOT be registered.
	removedFlags := []string{
		"debug", "debug-full", "namespace", "snapshotter", "cgroup-manager",
		"insecure-registry", "hosts-dir", "experimental", "host-gateway-ip",
		"kube-hide-dupe", "cdi-spec-dirs", "global-dns", "global-dns-opts", "global-dns-search",
	}
	for _, name := range removedFlags {
		assert.Nil(t, flags.Lookup(name), "flag %q should not be registered in finch-shim", name)
	}
}

func TestLoggingMagicArgv(t *testing.T) {
	assert.Equal(t, "_NERDCTL_INTERNAL_LOGGING", logging.MagicArgv1,
		"MagicArgv1 constant should match expected value used in run()")
}
