// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"
	"encoding/json"
	"fmt"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/containerinspector"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/containerd/nerdctl/v2/pkg/netutil"
	"github.com/containernetworking/cni/libcni"
	cnitypes "github.com/containernetworking/cni/pkg/types"
)

//go:generate mockgen --destination=../../mocks/mocks_backend/nerdctlnetworksvc.go -package=mocks_backend github.com/runfinch/finch-daemon/internal/backend NerdctlNetworkSvc
type NerdctlNetworkSvc interface {
	FilterNetworks(filterf func(networkConfig *netutil.NetworkConfig) bool) ([]*netutil.NetworkConfig, error)
	AddNetworkList(ctx context.Context, netconflist *libcni.NetworkConfigList, conf *libcni.RuntimeConf) (cnitypes.Result, error)
	CreateNetwork(opts types.NetworkCreateOptions) (*netutil.NetworkConfig, error)
	RemoveNetwork(networkConfig *netutil.NetworkConfig) error
	InspectNetwork(ctx context.Context, networkConfig *netutil.NetworkConfig) (*dockercompat.Network, error)
	UsedNetworkInfo(ctx context.Context) (map[string][]string, error)
	NetconfPath() string
	Namespace() string
}

func (w *NerdctlWrapper) FilterNetworks(filterf func(networkConfig *netutil.NetworkConfig) bool) ([]*netutil.NetworkConfig, error) {
	networkConfigs, err := w.netClient.NetworkList()
	if err != nil {
		return nil, err
	}
	result := []*netutil.NetworkConfig{}
	for _, networkConfig := range networkConfigs {
		if filterf(networkConfig) {
			result = append(result, networkConfig)
		}
	}
	return result, nil
}

func (w *NerdctlWrapper) AddNetworkList(ctx context.Context, netconflist *libcni.NetworkConfigList, conf *libcni.RuntimeConf) (cnitypes.Result, error) {
	return w.CNI.AddNetworkList(ctx, netconflist, conf)
}

func (w *NerdctlWrapper) CreateNetwork(opts types.NetworkCreateOptions) (*netutil.NetworkConfig, error) {
	return w.netClient.CreateNetwork(opts)
}

func (w *NerdctlWrapper) RemoveNetwork(networkConfig *netutil.NetworkConfig) error {
	return w.netClient.RemoveNetwork(networkConfig)
}

func (w *NerdctlWrapper) InspectNetwork(ctx context.Context, networkConfig *netutil.NetworkConfig) (*dockercompat.Network, error) {
	// Get containers associated with this network
	containers, err := getContainersFromNetConfig(ctx, networkConfig, w.clientWrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to get containers for network: %w", err)
	}

	network := &native.Network{
		CNI:           json.RawMessage(networkConfig.Bytes),
		NerdctlID:     networkConfig.NerdctlID,
		NerdctlLabels: networkConfig.NerdctlLabels,
		File:          networkConfig.File,
		Containers:    containers,
	}
	return dockercompat.NetworkFromNative(network)
}

func (w *NerdctlWrapper) UsedNetworkInfo(ctx context.Context) (map[string][]string, error) {
	return netutil.UsedNetworks(ctx, w.clientWrapper.client)
}

func (w *NerdctlWrapper) NetconfPath() string {
	return w.netClient.NetconfPath
}

func (w *NerdctlWrapper) Namespace() string {
	return w.netClient.Namespace
}

// getContainersFromNetConfig returns containers associated with the given network.
func getContainersFromNetConfig(ctx context.Context, networkConfig *netutil.NetworkConfig, client ContainerdClient) ([]*native.Container, error) {
	filters := []string{fmt.Sprintf(`labels.%q~="\\\"%s\\\""`, labels.Networks, networkConfig.Name)}
	filteredContainers, err := client.GetContainers(ctx, filters...)
	if err != nil {
		return nil, err
	}

	var containers []*native.Container
	for _, container := range filteredContainers {
		nativeContainer, err := containerinspector.Inspect(ctx, container)
		if err != nil {
			continue
		}
		if nativeContainer.Process == nil || nativeContainer.Process.Status.Status != containerd.Running {
			continue
		}

		containers = append(containers, nativeContainer)
	}

	return containers, nil
}
