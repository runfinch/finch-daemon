// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package statsutil

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	dockertypes "github.com/docker/docker/api/types/container"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

//go:generate mockgen --destination=../../mocks/mocks_statsutil/statsutil.go -package=mocks_statsutil github.com/runfinch/finch-daemon/pkg/statsutil StatsUtil
type StatsUtil interface {
	// GetSystemCPUUsage returns the host system's cpu usage in
	// nanoseconds. An error is returned if the format of the underlying
	// file does not match.
	GetSystemCPUUsage() (uint64, error)

	// GetNumberOnlineCPUs estimates number of available CPUs
	GetNumberOnlineCPUs() (uint32, error)

	// CollectNetworkStats collects network usage statistics for specified network
	// interfaces in a process namespace
	CollectNetworkStats(pid int, interfaces []native.NetInterface) (map[string]dockertypes.NetworkStats, error)
}

const (
	// From https://github.com/moby/moby/blob/v24.0.2/daemon/stats/collector_unix.go#L20-L21
	clockTicksPerSecond  = 100
	nanoSecondsPerSecond = 1e9
)

type statsUtil struct{}

func NewStatsUtil() StatsUtil {
	return &statsUtil{}
}

// GetSystemCPUUsage returns the host system's cpu usage in
// nanoseconds. An error is returned if the format of the underlying
// file does not match.
//
// Uses /proc/stat defined by POSIX. Looks for the cpu
// statistics line and then sums up the first seven fields
// provided. See `man 5 proc` for details on specific field
// information.
//
// Adapted from https://github.com/moby/moby/blob/v24.0.2/daemon/stats/collector_unix.go#L24-L67
func (s *statsUtil) GetSystemCPUUsage() (uint64, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	bufReader := bufio.NewReader(f)

	for {
		line, err := bufReader.ReadString('\n')
		if err != nil {
			break
		}
		parts := strings.Fields(line)
		switch parts[0] {
		case "cpu":
			if len(parts) < 8 {
				return 0, fmt.Errorf("invalid number of cpu fields")
			}
			var totalClockTicks uint64
			for _, i := range parts[1:8] {
				v, err := strconv.ParseUint(i, 10, 64)
				if err != nil {
					return 0, fmt.Errorf("unable to convert value %s to int: %s", i, err)
				}
				totalClockTicks += v
			}
			return (totalClockTicks * nanoSecondsPerSecond) /
				clockTicksPerSecond, nil
		}
	}
	return 0, fmt.Errorf("invalid stat format. Error trying to parse the '/proc/stat' file")
}

// Adapted from https://github.com/moby/moby/blob/v24.0.2/daemon/stats/collector_unix.go#L69-L76
func (s *statsUtil) GetNumberOnlineCPUs() (uint32, error) {
	var cpuset unix.CPUSet
	err := unix.SchedGetaffinity(0, &cpuset)
	if err != nil {
		return 0, err
	}
	return uint32(cpuset.Count()), nil
}

func (s *statsUtil) CollectNetworkStats(pid int, interfaces []native.NetInterface) (map[string]dockertypes.NetworkStats, error) {
	// get network namespace of the process
	ns, err := netns.GetFromPid(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to get network namespace from pid %d: %s", pid, err)
	}
	defer ns.Close()
	nlHandle, err := netlink.NewHandleAt(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the statistics in netns %s: %s", ns, err)
	}
	defer nlHandle.Close()

	// collect network stats for each network interface
	networks := map[string]dockertypes.NetworkStats{}
	for _, v := range interfaces {
		nlink, err := nlHandle.LinkByIndex(v.Index)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve the statistics for %s in netns %s: %s", v.Name, ns, err)
		}

		if nlink.Attrs().Flags&net.FlagUp != 0 {
			// exclude loopback interface
			if nlink.Attrs().Flags&net.FlagLoopback != 0 || strings.HasPrefix(nlink.Attrs().Name, "lo") {
				continue
			}

			net := dockertypes.NetworkStats{}
			stats := nlink.Attrs().Statistics
			if stats != nil {
				net.RxBytes = stats.RxBytes
				net.TxBytes = stats.TxBytes
				net.RxDropped = stats.RxDropped
				net.TxDropped = stats.TxDropped
				net.RxErrors = stats.RxErrors
				net.TxErrors = stats.TxErrors
				net.RxPackets = stats.RxPackets
				net.TxPackets = stats.TxPackets
			}
			networks[v.Name] = net
		}
	}

	return networks, nil
}
