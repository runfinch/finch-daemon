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
	"github.com/moby/moby/api/types/blkiodev"
	specs "github.com/opencontainers/runtime-spec/specs-go"
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

	// Enrich HostConfig with fields not available in dockercompat.HostConfig
	// These are extracted from the OCI spec and container labels.
	if cont.HostConfig != nil {
		spec, err := c.Spec(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get container spec: %s", err)
		}
		enrichHostConfigFromSpec(cont.HostConfig, spec, l)
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

	// Convert blkio weight devices from dockercompat to moby types
	var blkioWeightDevices []*blkiodev.WeightDevice
	for _, wd := range c.BlkioWeightDevice {
		if wd != nil {
			blkioWeightDevices = append(blkioWeightDevices, &blkiodev.WeightDevice{
				Path:   wd.Path,
				Weight: wd.Weight,
			})
		}
	}

	// Convert blkio throttle devices from dockercompat to moby types
	convertThrottleDevices := func(devices []*dockercompat.ThrottleDevice) []*blkiodev.ThrottleDevice {
		var result []*blkiodev.ThrottleDevice
		for _, td := range devices {
			if td != nil {
				result = append(result, &blkiodev.ThrottleDevice{
					Path: td.Path,
					Rate: td.Rate,
				})
			}
		}
		return result
	}

	return &types.ContainerHostConfig{
		ContainerIDFile: c.ContainerIDFile,
		LogConfig: types.LogConfig{
			Type:   c.LogConfig.Driver,
			Config: c.LogConfig.Opts,
		},
		PortBindings:         c.PortBindings,
		IpcMode:              c.IpcMode,
		PidMode:              c.PidMode,
		ReadonlyRootfs:       c.ReadonlyRootfs,
		ShmSize:              c.ShmSize,
		Sysctls:              c.Sysctls,
		CPUSetMems:           c.CPUSetMems,
		CPUSetCPUs:           c.CPUSetCPUs,
		CPUShares:            int64(c.CPUShares),
		CPUPeriod:            int64(c.CPUPeriod),
		CPUQuota:             c.CPUQuota,
		Memory:               c.Memory,
		MemorySwap:           c.MemorySwap,
		OomKillDisable:       c.OomKillDisable,
		Devices:              hostConfigDevices,
		CgroupnsMode:         types.CgroupnsMode(c.CgroupnsMode),
		DNS:                  c.DNS,
		DNSOptions:           c.DNSOptions,
		DNSSearch:            c.DNSSearch,
		ExtraHosts:           c.ExtraHosts,
		GroupAdd:             c.GroupAdd,
		Tmpfs:                c.Tmpfs,
		UTSMode:              c.UTSMode,
		Runtime:              c.Runtime,
		BlkioWeight:          c.BlkioWeight,
		BlkioWeightDevice:    blkioWeightDevices,
		BlkioDeviceReadBps:   convertThrottleDevices(c.BlkioDeviceReadBps),
		BlkioDeviceWriteBps:  convertThrottleDevices(c.BlkioDeviceWriteBps),
		BlkioDeviceReadIOps:  convertThrottleDevices(c.BlkioDeviceReadIOps),
		BlkioDeviceWriteIOps: convertThrottleDevices(c.BlkioDeviceWriteIOps),
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

// defaultCaps is the set of capabilities granted to a container by default,
// matching containerd's defaultUnixCaps(). Used to compute CapAdd/CapDrop deltas.
var defaultCaps = map[string]struct{}{
	"CAP_CHOWN":            {},
	"CAP_DAC_OVERRIDE":     {},
	"CAP_FSETID":           {},
	"CAP_FOWNER":           {},
	"CAP_MKNOD":            {},
	"CAP_NET_RAW":          {},
	"CAP_SETGID":           {},
	"CAP_SETUID":           {},
	"CAP_SETFCAP":          {},
	"CAP_SETPCAP":          {},
	"CAP_NET_BIND_SERVICE": {},
	"CAP_SYS_CHROOT":       {},
	"CAP_KILL":             {},
	"CAP_AUDIT_WRITE":      {},
}

// enrichHostConfigFromSpec populates HostConfig fields that are not available
// in the upstream dockercompat.HostConfig struct. These fields are extracted
// from the OCI spec and container labels.
func enrichHostConfigFromSpec(hc *types.ContainerHostConfig, spec *specs.Spec, containerLabels map[string]string) {
	if spec == nil {
		return
	}

	// Extract capabilities from OCI spec.
	// CapAdd = caps in bounding set that are NOT in the default set.
	// CapDrop = caps in the default set that are NOT in the bounding set.
	if spec.Process != nil && spec.Process.Capabilities != nil {
		caps := spec.Process.Capabilities

		// Detect privileged mode: if bounding set has all known capabilities
		hc.Privileged = isPrivileged(caps.Bounding)

		bounding := make(map[string]struct{}, len(caps.Bounding))
		for _, c := range caps.Bounding {
			bounding[c] = struct{}{}
		}

		for _, c := range caps.Bounding {
			if _, isDefault := defaultCaps[c]; !isDefault {
				hc.CapAdd = append(hc.CapAdd, c)
			}
		}
		for c := range defaultCaps {
			if _, present := bounding[c]; !present {
				hc.CapDrop = append(hc.CapDrop, c)
			}
		}
	}

	// Extract PidsLimit from OCI spec
	if spec.Linux != nil && spec.Linux.Resources != nil && spec.Linux.Resources.Pids != nil && spec.Linux.Resources.Pids.Limit != nil {
		hc.PidsLimit = *spec.Linux.Resources.Pids.Limit
	}

	// Extract Ulimits from OCI spec rlimits
	if spec.Process != nil && len(spec.Process.Rlimits) > 0 {
		for _, rl := range spec.Process.Rlimits {
			// Convert OCI rlimit type (e.g., "RLIMIT_NOFILE") to Docker format (e.g., "nofile")
			name := strings.TrimPrefix(rl.Type, "RLIMIT_")
			name = strings.ToLower(name)
			hc.Ulimits = append(hc.Ulimits, &types.Ulimit{
				Name: name,
				Hard: int64(rl.Hard),
				Soft: int64(rl.Soft),
			})
		}
	}

	// Extract annotations from OCI spec (filter out internal nerdctl labels)
	if len(spec.Annotations) > 0 {
		annotations := make(map[string]string)
		for k, v := range spec.Annotations {
			if !strings.HasPrefix(k, "nerdctl/") {
				annotations[k] = v
			}
		}
		if len(annotations) > 0 {
			hc.Annotations = annotations
		}
	}

	// Extract NetworkMode from container labels
	if networksJSON, ok := containerLabels[labels.Networks]; ok {
		var networks []string
		if err := json.Unmarshal([]byte(networksJSON), &networks); err == nil && len(networks) > 0 {
			hc.NetworkMode = networks[0]
		}
	}

	// Extract AutoRemove from container labels
	if autoRemove, ok := containerLabels[labels.ContainerAutoRemove]; ok {
		hc.AutoRemove, _ = strconv.ParseBool(autoRemove)
	}

	// Extract Binds from container labels (nerdctl/mounts)
	if mountsJSON, ok := containerLabels[labels.Mounts]; ok {
		var mounts []mountInfo
		if err := json.Unmarshal([]byte(mountsJSON), &mounts); err == nil {
			for _, m := range mounts {
				if m.Type == "bind" {
					bind := m.Source + ":" + m.Destination
					if m.Mode != "" {
						bind += ":" + m.Mode
					}
					hc.Binds = append(hc.Binds, bind)
				}
			}
		}
	}

	// Extract SecurityOpt from OCI spec
	if spec.Process != nil && spec.Process.ApparmorProfile != "" {
		hc.SecurityOpt = append(hc.SecurityOpt, "apparmor="+spec.Process.ApparmorProfile)
	}
	if spec.Process != nil && spec.Process.SelinuxLabel != "" {
		hc.SecurityOpt = append(hc.SecurityOpt, "label="+spec.Process.SelinuxLabel)
	}

	// Extract Init from OCI spec — check if init process is configured
	if spec.Process != nil && len(spec.Process.Args) > 0 {
		// nerdctl uses tini as init; check if the entrypoint is an init binary
		arg0 := spec.Process.Args[0]
		if strings.HasSuffix(arg0, "tini") || strings.HasSuffix(arg0, "docker-init") {
			initTrue := true
			hc.Init = &initTrue
		}
	}
}

// mountInfo is a minimal struct for parsing nerdctl mount labels.
type mountInfo struct {
	Type        string `json:"Type"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Mode        string `json:"Mode"`
}

// isPrivileged checks if the bounding capability set contains all known capabilities.
// This is a heuristic — a container is considered privileged if it has a very large set of capabilities.
func isPrivileged(boundingCaps []string) bool {
	// A privileged container typically has 40+ capabilities.
	// The exact number varies by kernel version, but 38+ is a strong signal.
	return len(boundingCaps) >= 38
}
