// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/runfinch/finch-daemon/api/types"
)

const networkPrefix = "unknown-eth"

func (s *service) Inspect(ctx context.Context, cid string, sizeFlag bool) (*types.Container, error) {
	c, err := s.getContainer(ctx, cid)
	if err != nil {
		return nil, err
	}

	inspect, err := s.nctlContainerSvc.InspectContainer(ctx, c, sizeFlag)
	if err != nil {
		return nil, err
	}

	// translate to a finch-daemon container inspect type
	cont := types.Container{
		ID:              inspect.ID,
		Created:         inspect.Created,
		Path:            inspect.Path,
		Args:            inspect.Args,
		State:           inspect.State,
		Image:           inspect.Image,
		ResolvConfPath:  inspect.ResolvConfPath,
		HostnamePath:    inspect.HostnamePath,
		LogPath:         inspect.LogPath,
		Name:            fmt.Sprintf("/%s", inspect.Name),
		RestartCount:    inspect.RestartCount,
		Driver:          inspect.Driver,
		Platform:        inspect.Platform,
		AppArmorProfile: inspect.AppArmorProfile,
		Mounts:          inspect.Mounts,
		NetworkSettings: inspect.NetworkSettings,
		SizeRw:          inspect.SizeRw,
		SizeRootFs:      inspect.SizeRootFs,
	}

	cont.Config = &types.ContainerConfig{
		Hostname:     inspect.Config.Hostname,
		User:         inspect.Config.User,
		AttachStdin:  inspect.Config.AttachStdin,
		ExposedPorts: inspect.Config.ExposedPorts,
		Tty:          false, // TODO: Tty is always false until attach supports stdin with tty
		Env:          inspect.Config.Env,
		Cmd:          inspect.Config.Cmd,
		Image:        inspect.Image,
		Volumes:      inspect.Config.Volumes,
		WorkingDir:   inspect.Config.WorkingDir,
		Entrypoint:   inspect.Config.Entrypoint,
		Labels:       inspect.Config.Labels,
	}

	cont.HostConfig = getHostConfigFromDockerCompat(inspect.HostConfig)

	l, err := c.Labels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container labels: %s", err)
	}
	updateNetworkSettings(ctx, cont.NetworkSettings, l)

	// make sure it passes the default time value for time fields otherwise the goclient fails.
	if inspect.Created == "" {
		cont.Created = "0001-01-01T00:00:00Z"
	}

	if inspect.State != nil && inspect.State.FinishedAt == "" {
		cont.State.FinishedAt = "0001-01-01T00:00:00Z"
	}

	return &cont, nil
}

func getHostConfigFromDockerCompat(c *dockercompat.HostConfig) *types.ContainerHostConfig {
	if c == nil {
		return nil
	}

	hostConfigDevices := []types.DeviceMapping{}
	for _, device := range c.Devices {
		hostConfigDevices = append(hostConfigDevices, types.DeviceMapping{
			PathOnHost:        device.PathOnHost,
			PathInContainer:   device.PathInContainer,
			CgroupPermissions: device.CgroupPermissions,
		})
	}

	return &types.ContainerHostConfig{
		ContainerIDFile: c.ContainerIDFile,
		LogConfig: types.LogConfig{
			Type:   c.LogConfig.Driver,
			Config: c.LogConfig.Opts,
		},
		PortBindings:   c.PortBindings,
		IpcMode:        c.IpcMode,
		PidMode:        c.PidMode,
		ReadonlyRootfs: c.ReadonlyRootfs,
		ShmSize:        c.ShmSize,
		Sysctls:        c.Sysctls,
		CPUSetMems:     c.CPUSetMems,
		CPUSetCPUs:     c.CPUSetCPUs,
		CPUShares:      int64(c.CPUShares),
		CPUPeriod:      int64(c.CPUPeriod),
		Memory:         c.Memory,
		MemorySwap:     c.MemorySwap,
		OomKillDisable: c.OomKillDisable,
		Devices:        hostConfigDevices,
		// TODO: Uncomment these when https://github.com/runfinch/finch-daemon/pull/267 gets merged
		// CPURealtimePeriod: inspect.HostConfig.CPURealtimePeriod,
		// CPURealtimeRuntime: inspect.HostConfig.CPURealtimeRuntime,
		// TODO: add blkio devices which requires a change in the dockercompat response from Nerdctl
	}
}

// updateNetworkSettings updates the settings in the network to match that
// of docker as docker identifies networks by their name in "NetworkSettings",
// but nerdctl uses a sequential ordering "unknown-eth0", "unknown-eth1",...
// we use container labels to find corresponding name for each network in "NetworkSettings".
func updateNetworkSettings(ctx context.Context, ns *dockercompat.NetworkSettings, labels map[string]string) error {
	if ns != nil && ns.Networks != nil {
		networks := map[string]*dockercompat.NetworkEndpointSettings{}

		for network, settings := range ns.Networks {
			networkName := getNetworkName(labels, network)
			networks[networkName] = settings
		}
		ns.Networks = networks
	}
	return nil
}

// getNetworkName gets network name from container labels using the index specified by the network prefix.
// returns the default prefix if network name was not found.
func getNetworkName(lab map[string]string, network string) string {
	namesJSON, ok := lab[labels.Networks]
	if !ok {
		return network
	}
	var names []string
	if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
		return network
	}

	if strings.HasPrefix(network, networkPrefix) {
		prefixLen := len(networkPrefix)
		index, err := strconv.ParseUint(network[prefixLen:], 10, 64)
		if err != nil {
			return network
		}
		if int(index) < len(names) {
			return names[index]
		}
	}

	return network
}
