// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// PoC: Solution IV — decouple CNI networking from OCI hooks.
// Instead of relying on the nerdctl OCI hook binary to call cni.Setup() at
// createRuntime time, finch-daemon pre-creates a named network namespace and
// calls cni.Setup() directly during container create.  The netns path is then
// injected into the OCI spec so runc joins the pre-existing namespace rather
// than creating a new one.

package backend

import (
	"context"
	"fmt"

	gocni "github.com/containerd/go-cni"
	"github.com/containerd/nerdctl/v2/pkg/netutil"
	"github.com/vishvananda/netns"
)

// ContainerCNISvc is the interface for CNI + netns lifecycle operations.
// The interface itself is defined in cni_netns_other.go (shared across all
// platforms) so that the mock generator and non-Linux builds can reference it.
// Linux provides the real implementations below.

// netnsName returns a deterministic named-netns identifier for a container.
func netnsName(containerID string) string {
	// Keep it short; named netns live under /var/run/netns/<name>.
	if len(containerID) > 12 {
		return "finch-" + containerID[:12]
	}
	return "finch-" + containerID
}

// SetupContainerNetwork implements ContainerCNISvc.
func (w *NerdctlWrapper) SetupContainerNetwork(ctx context.Context, containerID, networkName string) (string, error) {
	name := netnsName(containerID)

	// Create a new named network namespace.  If one already exists (e.g. a
	// previous failed attempt) delete it first so we start clean.
	_ = netns.DeleteNamed(name) // best-effort; ignore error

	ns, err := netns.NewNamed(name)
	if err != nil {
		return "", fmt.Errorf("failed to create named netns %q: %w", name, err)
	}
	ns.Close() // we only need the path, not the fd

	nsPath := fmt.Sprintf("/var/run/netns/%s", name)

	// Build a go-cni instance for the requested network.
	cniInstance, err := w.buildCNIForNetwork(networkName)
	if err != nil {
		_ = netns.DeleteNamed(name)
		return "", err
	}

	fullID := w.globalOptions.Namespace + "-" + containerID
	if _, err := cniInstance.Setup(ctx, fullID, nsPath); err != nil {
		_ = netns.DeleteNamed(name)
		return "", fmt.Errorf("cni.Setup failed: %w", err)
	}

	return nsPath, nil
}

// RemoveContainerNetwork implements ContainerCNISvc.
func (w *NerdctlWrapper) RemoveContainerNetwork(ctx context.Context, containerID, networkName, nsPath string) error {
	cniInstance, err := w.buildCNIForNetwork(networkName)
	if err != nil {
		return err
	}

	fullID := w.globalOptions.Namespace + "-" + containerID
	// Pass empty nsPath to Remove — the network namespace may already be gone.
	if err := cniInstance.Remove(ctx, fullID, ""); err != nil {
		return fmt.Errorf("cni.Remove failed: %w", err)
	}

	name := netnsName(containerID)
	if err := netns.DeleteNamed(name); err != nil {
		return fmt.Errorf("failed to delete named netns %q: %w", name, err)
	}
	return nil
}

// buildCNIForNetwork constructs a go-cni instance configured for networkName.
func (w *NerdctlWrapper) buildCNIForNetwork(networkName string) (gocni.CNI, error) {
	e, err := netutil.NewCNIEnv(
		w.globalOptions.CNIPath,
		w.globalOptions.CNINetConfPath,
		netutil.WithNamespace(w.globalOptions.Namespace),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CNI env: %w", err)
	}

	netw, err := e.NetworkByNameOrID(networkName)
	if err != nil {
		return nil, fmt.Errorf("network %q not found: %w", networkName, err)
	}

	cniInstance, err := gocni.New(
		gocni.WithPluginDir([]string{w.globalOptions.CNIPath}),
		gocni.WithConfListBytes(netw.Bytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create go-cni instance: %w", err)
	}
	return cniInstance, nil
}
