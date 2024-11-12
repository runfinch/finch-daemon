// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/containerd/containerd"
	cerrdefs "github.com/containerd/errdefs"
	gocni "github.com/containerd/go-cni"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/containerd/nerdctl/v2/pkg/netutil"
	"github.com/containerd/nerdctl/v2/pkg/netutil/nettype"
	"github.com/containerd/nerdctl/v2/pkg/strutil"
	"github.com/containernetworking/cni/libcni"
	"github.com/sirupsen/logrus"
)

const (
	// go-cni default
	// From https://github.com/containerd/go-cni/blob/31de2455ae5d8bfc743572bfe73587ca7468f865/types.go
	interfacePrefix = "eth"

	// This container label stores the current maximum network index.
	// Eg: if the network index is 1, a new network will be added with interface name eth2 and the index will be updated.
	// Maintaining this index is important as opposed to just using the length of networks as the index,
	// because the index may be different from the length if a network was removed using the network disconnect API.
	networkIndexLabel = "finch/network-index"
)

func (s *service) Connect(ctx context.Context, networkId, containerId string) error {
	logrus.Infof("network connect: network Id %s, container Id %s", networkId, containerId)

	net, err := s.getNetwork(networkId)
	if err != nil {
		logrus.Debugf("Failed to get network: %s", err)
		return err
	}
	container, err := s.getContainer(ctx, containerId)
	if err != nil {
		logrus.Debugf("Failed to get container: %s", err)
		return err
	}
	status, err := getContainerStatus(ctx, container)
	if err != nil {
		logrus.Debugf("Failed to get container status: %s", err)
		return err
	}

	switch status {
	case containerd.Unknown:
		return fmt.Errorf("failed to determine container status from runtime")
	case containerd.Pausing:
		return fmt.Errorf("cannot connect a container that is currently pausing")

	// the container runtime has been set up and network namespace already exists,
	// connect this network to the existing namespace for the running task
	case containerd.Paused, containerd.Running:
		deleteFunc, err := addNetworkConfig(ctx, net, container)
		if err != nil {
			return err
		}
		err = s.connectNetwork(ctx, net, container)
		if err != nil {
			// cleanup added network
			deleteFunc()
			return err
		}
		return nil

	// default is the case when the container is either stopped or created, i.e. a network namespace does not exist,
	// the new network configuration must be added to the existing list of networks
	default:
		_, err = addNetworkConfig(ctx, net, container)
		return err
	}
}

func addNetworkConfig(ctx context.Context, net *netutil.NetworkConfig, container containerd.Container) (func(), error) {
	opts, err := container.Labels(ctx)
	if err != nil {
		logrus.Errorf("Failed to get container labels: %s", err)
		return nil, err
	}
	networksJSON := opts[labels.Networks]
	var networks []string
	err = json.Unmarshal([]byte(networksJSON), &networks)
	if err != nil {
		logrus.Errorf("Failed to unmarshal networks json %s: %s", networksJSON, err)
		return nil, err
	}
	if strutil.InStringSlice(networks, net.Name) {
		logrus.Debugf("Container %s already connected to network %s", container.ID(), net.Name)
		return nil, fmt.Errorf("container %s already connected to network %s", container.ID(), net.Name)
	}

	err = verifyNetworkConfig(net, networks, opts[labels.MACAddress])
	if err != nil {
		logrus.Debugf("Failed to verify network configuration: %s", err)
		return nil, err
	}
	networks = append(networks, net.Name)

	// update OCI spec
	spec, err := container.Spec(ctx)
	if err != nil {
		logrus.Errorf("Failed to get container OCI spec: %s", err)
		return nil, err
	}
	networksData, err := json.Marshal(networks)
	if err != nil {
		logrus.Errorf("Failed to marshal networks slice %v: %s", networks, err)
		return nil, err
	}
	opts[labels.Networks] = string(networksData)
	index := -1
	if indexString, ok := opts[networkIndexLabel]; ok {
		index, err = strconv.Atoi(indexString)
		if err != nil {
			logrus.Errorf("Invalid network index %s: %s", indexString, err)
			return nil, err
		}
		index = index + 1
	} else {
		index = len(networks) - 1
	}
	opts[networkIndexLabel] = strconv.Itoa(index)
	spec.Annotations[labels.Networks] = string(networksData)
	err = container.Update(ctx,
		containerd.UpdateContainerOpts(containerd.WithContainerLabels(opts)),
		containerd.UpdateContainerOpts(containerd.WithSpec(spec)),
	)
	if err != nil {
		logrus.Errorf("Failed to update container: %s", err)
		return nil, err
	}

	// define a garbage collector to remove the added network from container spec,
	// to be used when the network cannot be attached successfully.
	deleteNet := func() {
		opts[networkIndexLabel] = strconv.Itoa(index - 1)
		networks = networks[:len(networks)-1]
		networksData, err = json.Marshal(networks)
		if err != nil {
			logrus.Errorf("Could not marshal networks slice %v: %s", networks, err)
			return
		}
		opts[labels.Networks] = string(networksData)
		spec.Annotations[labels.Networks] = string(networksData)
		err = container.Update(ctx,
			containerd.UpdateContainerOpts(containerd.WithContainerLabels(opts)),
			containerd.UpdateContainerOpts(containerd.WithSpec(spec)),
		)
		if err != nil {
			logrus.Errorf("Failed to update container: %s", err)
			return
		}
	}

	return deleteNet, nil
}

func (s *service) connectNetwork(ctx context.Context, net *netutil.NetworkConfig, container containerd.Container) error {
	nsPath, err := getContainerNetNSPath(ctx, container)
	if err != nil {
		logrus.Errorf("Failed to get container network namespace path: %s", err)
		return err
	}
	networkIndex, err := getNetworkIndex(ctx, container)
	if err != nil {
		logrus.Errorf("Failed to get container network index: %s", err)
		return err
	}

	// define CNI ADD configuration
	cniAddConfig := &libcni.RuntimeConf{
		ContainerID: container.ID(),
		NetNS:       nsPath,
		IfName:      interfacePrefix + strconv.Itoa(networkIndex),
	}
	opts, err := container.Labels(ctx)
	if err != nil {
		logrus.Errorf("Failed to get container labels: %s", err)
		return err
	}
	args := [][2]string{{"IgnoreUnknown", "1"}}
	if ipAddress, ok := opts[labels.IPAddress]; ok && ipAddress != "" {
		args = append(args, [2]string{"IP", ipAddress})
	}
	if macAddress, ok := opts[labels.MACAddress]; ok && macAddress != "" {
		args = append(args, [2]string{"MAC", macAddress})
	}
	if portsJSON, ok := opts[labels.Ports]; ok && portsJSON != "" {
		var ports []gocni.PortMapping
		if err := json.Unmarshal([]byte(portsJSON), &ports); err != nil {
			logrus.Errorf("Failed to unmarshal ports from labels: %s", err)
			return err
		}
		cniAddConfig.CapabilityArgs = make(map[string]interface{})
		cniAddConfig.CapabilityArgs["portMappings"] = ports
	}
	cniAddConfig.Args = args

	// attach network
	_, err = s.netClient.AddNetworkList(ctx, net.NetworkConfigList, cniAddConfig)
	if err != nil {
		logrus.Errorf("Error attaching network %s to container %s: %s", net.Name, container.ID(), err)
		return err
	}

	return nil
}

func getContainerStatus(ctx context.Context, container containerd.Container) (containerd.ProcessStatus, error) {
	task, err := container.Task(ctx, nil)
	if err != nil {
		// no running task found implies the container was created but never started
		if cerrdefs.IsNotFound(err) {
			return containerd.Created, nil
		}
		return containerd.Unknown, err
	}
	status, err := task.Status(ctx)
	if err != nil {
		return containerd.Unknown, err
	}
	return status.Status, nil
}

func getContainerNetNSPath(ctx context.Context, container containerd.Container) (string, error) {
	task, err := container.Task(ctx, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/proc/%d/ns/net", task.Pid()), nil
}

func getNetworkIndex(ctx context.Context, container containerd.Container) (int, error) {
	opts, err := container.Labels(ctx)
	if err != nil {
		return -1, err
	}
	if indexString, ok := opts[networkIndexLabel]; ok {
		index, err := strconv.Atoi(indexString)
		if err != nil {
			return -1, err
		}
		return index, nil
	}

	// return the length of existing networks otherwise
	networksJSON := opts[labels.Networks]
	var networks []string
	err = json.Unmarshal([]byte(networksJSON), &networks)
	if err != nil {
		return -1, err
	}
	opts[networkIndexLabel] = strconv.Itoa(len(networks) - 1)
	_, err = container.SetLabels(ctx, opts)
	if err != nil {
		return -1, err
	}
	return len(networks) - 1, nil
}

func verifyNetworkConfig(net *netutil.NetworkConfig, networks []string, macAddress string) error {
	netType, err := nettype.Detect(append(networks, net.Name))
	if err != nil {
		return err
	}
	if netType != nettype.CNI {
		return fmt.Errorf("invalid network %s, only CNI type is supported", net.Name)
	}
	if macAddress != "" {
		macValidNetworks := []string{"bridge", "macvlan"}
		netMode := net.Plugins[0].Network.Type
		if !strutil.InStringSlice(macValidNetworks, netMode) {
			return fmt.Errorf("network type %q is not supported when MAC address is specified, must be one of: %v", netMode, macValidNetworks)
		}
	}
	return nil
}
