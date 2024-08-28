// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"
	"encoding/json"

	"github.com/containerd/nerdctl/pkg/inspecttypes/dockercompat"
	"github.com/containerd/nerdctl/pkg/inspecttypes/native"
	"github.com/containerd/nerdctl/pkg/netutil"
	"github.com/containernetworking/cni/libcni"
	cnitypes "github.com/containernetworking/cni/pkg/types"
)

//go:generate mockgen --destination=../../mocks/mocks_backend/nerdctlnetworksvc.go -package=mocks_backend github.com/runfinch/finch-daemon/internal/backend NerdctlNetworkSvc
type NerdctlNetworkSvc interface {
	FilterNetworks(filterf func(networkConfig *netutil.NetworkConfig) bool) ([]*netutil.NetworkConfig, error)
	AddNetworkList(ctx context.Context, netconflist *libcni.NetworkConfigList, conf *libcni.RuntimeConf) (cnitypes.Result, error)
	CreateNetwork(opts netutil.CreateOptions) (*netutil.NetworkConfig, error)
	RemoveNetwork(networkConfig *netutil.NetworkConfig) error
	InspectNetwork(ctx context.Context, networkConfig *netutil.NetworkConfig) (*dockercompat.Network, error)
	UsedNetworkInfo(ctx context.Context) (map[string][]string, error)
	NetconfPath() string
}

func (w *NerdctlWrapper) FilterNetworks(filterf func(networkConfig *netutil.NetworkConfig) bool) ([]*netutil.NetworkConfig, error) {
	return w.netClient.FilterNetworks(filterf)
}

func (w *NerdctlWrapper) AddNetworkList(ctx context.Context, netconflist *libcni.NetworkConfigList, conf *libcni.RuntimeConf) (cnitypes.Result, error) {
	return w.CNI.AddNetworkList(ctx, netconflist, conf)
}

func (w *NerdctlWrapper) CreateNetwork(opts netutil.CreateOptions) (*netutil.NetworkConfig, error) {
	return w.netClient.CreateNetwork(opts)
}

func (w *NerdctlWrapper) RemoveNetwork(networkConfig *netutil.NetworkConfig) error {
	return w.netClient.RemoveNetwork(networkConfig)
}

func (w *NerdctlWrapper) InspectNetwork(ctx context.Context, networkConfig *netutil.NetworkConfig) (*dockercompat.Network, error) {
	network := &native.Network{
		CNI:           json.RawMessage(networkConfig.Bytes),
		NerdctlID:     networkConfig.NerdctlID,
		NerdctlLabels: networkConfig.NerdctlLabels,
		File:          networkConfig.File,
	}
	return dockercompat.NetworkFromNative(network)
}

func (w *NerdctlWrapper) UsedNetworkInfo(ctx context.Context) (map[string][]string, error) {
	return netutil.UsedNetworks(ctx, w.clientWrapper.client)
}

func (w *NerdctlWrapper) NetconfPath() string {
	return w.netClient.NetconfPath
}
