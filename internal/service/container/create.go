// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/clientutil"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/containerd/nerdctl/v2/pkg/logging"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// labelNetnsPath is the container label used to store the pre-created netns path
// so that remove.go can look it up for cleanup.
const labelNetnsPath = "finch-daemon/netns-path"

// labelNetworkName is the container label used to store the CNI network name
// so that remove.go can call cni.Remove with the right network.
const labelNetworkName = "finch-daemon/network-name"

func (s *service) Create(ctx context.Context, image string, cmd []string, createOpt types.ContainerCreateOptions, netOpt types.NetworkOptions) (cid string, err error) {
	// PoC (Solution IV): Do NOT set NerdctlCmd.  nerdctl will still call
	// withNerdctlOCIHook("", ...) and bake hooks into the spec, but we strip
	// them immediately in updateContainerMetadata below.
	//
	// Networking is handled by finch-daemon directly via pre-created netns +
	// cni.Setup(), bypassing the hook mechanism entirely.

	netManager, err := s.nctlContainerSvc.NewNetworkingOptionsManager(netOpt)
	if err != nil {
		logrus.Debugf("error creating network manager for the given network options: %s", err)
		return "", err
	}

	args := []string{image}
	args = append(args, cmd...)
	cont, gc, err := s.nctlContainerSvc.CreateContainer(ctx, args, netManager, createOpt)
	if err != nil {
		if gc != nil {
			gc()
		}
		logrus.Debugf("failed to create container: %s", err)

		switch {
		case cerrdefs.IsNotFound(err):
			return "", errdefs.NewNotFound(err)
		case cerrdefs.IsInvalidArgument(err):
			return "", errdefs.NewInvalidFormat(err)
		case cerrdefs.IsAlreadyExists(err):
			return "", errdefs.NewConflict(err)
		default:
			return "", err
		}
	}

	// Determine the network name to use for CNI setup.
	// Default to "bridge" when no explicit network is requested.
	networkName := "bridge"
	if len(netOpt.NetworkSlice) > 0 && netOpt.NetworkSlice[0] != "" {
		networkName = netOpt.NetworkSlice[0]
	}

	if err := updateContainerMetadata(ctx, createOpt, netOpt, cont, s, networkName); err != nil {
		// Best-effort cleanup of the containerd container on metadata failure.
		_ = s.nctlContainerSvc.RemoveContainer(ctx, cont, true, false)
		return "", err
	}

	return cont.ID(), nil
}

func updateContainerMetadata(
	ctx context.Context,
	createOpt types.ContainerCreateOptions,
	netOpt types.NetworkOptions,
	cont containerd.Container,
	s *service,
	networkName string,
) error {
	// get container labels
	opts, err := cont.Labels(ctx)
	if err != nil {
		logrus.Errorf("failed to get container labels: %s", err)
		return err
	}
	// get oci spec
	spec, err := cont.Spec(ctx)
	if err != nil {
		logrus.Errorf("failed to get container OCI spec: %s", err)
		return err
	}

	// Handle log URI reset
	// NOTE: this is a temporary workaround to fix logging issue described in https://github.com/containerd/nerdctl/issues/2264.
	// The refactored create method in nerdctl uses self exe (finch-daemon) binary for logging instead of nerdctl binary path.
	// The following workaround resets this logging binary in the OCI spec and handles port labels for backward compatibility.
	// TODO: remove this workaround when the issue is resolved upstream.
	dataStore, err := clientutil.DataStore(createOpt.GOptions.DataRoot, createOpt.GOptions.Address)
	if err != nil {
		logrus.Errorf("failed to get nerdctl data store: %s", err)
		return err
	}

	ncExe, err := s.nctlContainerSvc.GetNerdctlExe()
	if err != nil {
		logrus.Errorf("failed to find nerdctl binary for log URI: %s", err)
		return err
	}

	logArgs := map[string]string{
		logging.MagicArgv1: dataStore,
	}
	logURI, err := cio.LogURIGenerator("binary", ncExe, logArgs)
	if err != nil {
		logrus.Errorf("failed to generate a log URI: %s", err)
		return err
	}

	opts[labels.LogURI] = logURI.String()
	spec.Annotations[labels.LogURI] = logURI.String()

	// Handle port labels for backward compatibility with nerdctl 2.1.2.
	if len(netOpt.PortMappings) > 0 {
		portsJSON, err := json.Marshal(netOpt.PortMappings)
		if err != nil {
			return err
		}
		opts[labels.Ports] = string(portsJSON)
		spec.Annotations[labels.Ports] = string(portsJSON)
	}

	// PoC (Solution IV): Strip OCI hooks that nerdctl baked in.
	// On Linux, nerdctl unconditionally injects createRuntime + postStop hooks
	// (via withNerdctlOCIHook) whenever NerdctlCmd != "" and GOOS != "windows".
	// Since we left NerdctlCmd empty the hook path will be "", which would fail
	// at runtime.  We remove them here and handle networking ourselves below.
	if runtime.GOOS == "linux" && spec.Hooks != nil {
		spec.Hooks.CreateRuntime = nil
		spec.Hooks.Poststop = nil
	}

	// PoC (Solution IV): Pre-create a named netns and run CNI setup.
	// We only do this for non-host, non-none network modes.
	var netnsPath string
	if runtime.GOOS == "linux" && networkName != "host" && networkName != "none" {
		netnsPath, err = s.nctlContainerSvc.SetupContainerNetwork(ctx, cont.ID(), networkName)
		if err != nil {
			logrus.Errorf("failed to setup container network: %s", err)
			return fmt.Errorf("failed to setup container network: %w", err)
		}

		// Inject the pre-created netns into the OCI spec so runc joins it.
		// We find the existing "network" namespace entry and replace its path.
		injected := false
		for i, ns := range spec.Linux.Namespaces {
			if ns.Type == "network" {
				spec.Linux.Namespaces[i].Path = netnsPath
				injected = true
				break
			}
		}
		if !injected {
			// No network namespace entry found — add one.
			spec.Linux.Namespaces = append(spec.Linux.Namespaces, specs.LinuxNamespace{
				Type: specs.NetworkNamespace,
				Path: netnsPath,
			})
		}

		// Store netns path and network name in labels for cleanup in Remove.
		opts[labelNetnsPath] = netnsPath
		opts[labelNetworkName] = networkName
	}

	err = cont.Update(ctx,
		containerd.UpdateContainerOpts(containerd.WithContainerLabels(opts)),
		containerd.UpdateContainerOpts(containerd.WithSpec(spec)),
	)
	if err != nil {
		// If we already set up the network, tear it down before returning.
		if netnsPath != "" {
			_ = s.nctlContainerSvc.RemoveContainerNetwork(ctx, cont.ID(), networkName, netnsPath)
		}
		logrus.Errorf("failed to update container: %s", err)
		return err
	}

	return nil
}
