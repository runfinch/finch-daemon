// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/containerd/cgroups/v3/cgroup1/stats"
	v2 "github.com/containerd/cgroups/v3/cgroup2/stats"
	containerd "github.com/containerd/containerd/v2/client"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/containerd/typeurl/v2"
	dockertypes "github.com/docker/docker/api/types/container"

	"github.com/runfinch/finch-daemon/api/types"
)

func (s *service) Stats(ctx context.Context, cid string) (<-chan *types.StatsJSON, error) {
	con, err := s.getContainer(ctx, cid)
	if err != nil {
		return nil, err
	}

	// get container name
	lab, err := con.Labels(ctx)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("/%s", lab[labels.Name])

	// listen to remove event for this container
	remove, removeErr := s.client.GetContainerRemoveEvent(ctx, con)

	statsCh := make(chan *types.StatsJSON, 100)
	ticker := time.NewTicker(time.Second)
	preStats := &types.StatsJSON{} // previous container stats data

	// start a goroutine to collect stats every second
	// until either the container is removed or the context is cancelled
	go func() {
		defer close(statsCh)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				statsJSON, err := s.collectContainerStats(ctx, con)
				if err != nil {
					// log warning and send an empty stats object
					s.logger.Warnf("error collecting container %s stats: %s", con.ID(), err)
					preStats = &types.StatsJSON{ID: con.ID(), Name: name}
					statsCh <- preStats
				} else if statsJSON == nil {
					// send an empty stats object
					preStats = &types.StatsJSON{ID: con.ID(), Name: name}
					statsCh <- preStats
				} else {
					// set current stats properties and update previous stats
					statsJSON.Read = time.Now()
					statsJSON.PreRead = preStats.Read
					statsJSON.PreCPUStats = preStats.CPUStats
					statsJSON.ID = con.ID()
					statsJSON.Name = name
					statsCh <- statsJSON
					preStats = statsJSON
				}
			case err = <-removeErr:
				if err != nil {
					s.logger.Errorf("container remove event error: %s", err)
				}
				return
			case <-remove:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return statsCh, nil
}

func (s *service) collectContainerStats(ctx context.Context, con containerd.Container) (*types.StatsJSON, error) {
	task, err := con.Task(ctx, nil)
	if err != nil {
		// if task was not found, it implies the container is not running, but it's not necessarily an error
		if cerrdefs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// return an empty stats object if the container is not running
	taskStatus, err := task.Status(ctx)
	if err != nil {
		return nil, err
	}
	status := taskStatus.Status
	if status == containerd.Created ||
		status == containerd.Stopped ||
		status == containerd.Unknown {
		return nil, nil
	}

	// get container cgroup metrics
	metrics, err := task.Metrics(ctx)
	if err != nil {
		return nil, err
	}
	anydata, err := typeurl.UnmarshalAny(metrics.Data)
	if err != nil {
		return nil, err
	}

	var st *types.StatsJSON
	switch v := anydata.(type) {
	case *v1.Metrics:
		st = collectCgroup1Stats(v)
	case *v2.Metrics:
		st = collectCgroup2Stats(v)
	default:
		return nil, fmt.Errorf("cannot convert metric data to cgroups.Metrics")
	}

	// get total system usage and number of cores
	systemUsage, err := s.stats.GetSystemCPUUsage()
	if err != nil {
		return nil, err
	}
	onlineCPUs, err := s.stats.GetNumberOnlineCPUs()
	if err != nil {
		return nil, err
	}
	st.CPUStats.SystemUsage = systemUsage
	st.CPUStats.OnlineCPUs = onlineCPUs

	// get network usage stats
	pid := int(task.Pid())
	netNS, err := s.nctlContainerSvc.InspectNetNS(ctx, pid)
	if err != nil {
		return nil, err
	}
	networks, err := s.stats.CollectNetworkStats(pid, netNS.Interfaces)
	if err != nil {
		return nil, err
	}
	if len(networks) > 0 {
		st.Networks = networks
	}

	return st, nil
}

// collectCgroup1Stats uses the cgroup v1 API to infer
// resource usage statistics from the metrics.
//
// Adapted from https://github.com/moby/moby/blob/v24.0.4/daemon/stats_unix.go#L57-L149
func collectCgroup1Stats(data *v1.Metrics) *types.StatsJSON {
	st := types.StatsJSON{}

	if data.Pids != nil {
		st.PidsStats = dockertypes.PidsStats{
			Current: data.Pids.Current,
			Limit:   data.Pids.Limit,
		}
	}

	if data.CPU != nil && data.CPU.Usage != nil {
		st.CPUStats = types.CPUStats{
			CPUUsage: dockertypes.CPUUsage{
				TotalUsage:        data.CPU.Usage.Total,
				PercpuUsage:       data.CPU.Usage.PerCPU,
				UsageInKernelmode: data.CPU.Usage.Kernel,
				UsageInUsermode:   data.CPU.Usage.User,
			},
			OnlineCPUs: uint32(len(data.CPU.Usage.PerCPU)),
		}
	}

	if data.Memory != nil && data.Memory.Usage != nil {
		st.MemoryStats = dockertypes.MemoryStats{
			Usage:    data.Memory.Usage.Usage,
			MaxUsage: data.Memory.Usage.Max,
			Failcnt:  data.Memory.Usage.Failcnt,
			Limit:    data.Memory.Usage.Limit,
		}
	}

	if data.Blkio != nil {
		st.BlkioStats = dockertypes.BlkioStats{
			IoServiceBytesRecursive: translateBlkioEntry(data.Blkio.IoServiceBytesRecursive),
			IoServicedRecursive:     translateBlkioEntry(data.Blkio.IoServicedRecursive),
			IoQueuedRecursive:       translateBlkioEntry(data.Blkio.IoQueuedRecursive),
			IoServiceTimeRecursive:  translateBlkioEntry(data.Blkio.IoServiceTimeRecursive),
			IoWaitTimeRecursive:     translateBlkioEntry(data.Blkio.IoWaitTimeRecursive),
			IoMergedRecursive:       translateBlkioEntry(data.Blkio.IoMergedRecursive),
			IoTimeRecursive:         translateBlkioEntry(data.Blkio.IoTimeRecursive),
			SectorsRecursive:        translateBlkioEntry(data.Blkio.SectorsRecursive),
		}
	}

	return &st
}

// collectCgroup2Stats uses the newer cgroup v2 API to infer
// resource usage statistics from the metrics
//
// Adapted from https://github.com/moby/moby/blob/v24.0.4/daemon/stats_unix.go#L151-L251
func collectCgroup2Stats(data *v2.Metrics) *types.StatsJSON {
	st := types.StatsJSON{}

	if data.Pids != nil {
		st.PidsStats = dockertypes.PidsStats{
			Current: data.Pids.Current,
			Limit:   data.Pids.Limit,
		}
	}

	if data.CPU != nil {
		st.CPUStats = types.CPUStats{
			CPUUsage: dockertypes.CPUUsage{
				TotalUsage: data.CPU.UsageUsec * 1000,
				// PercpuUsage is not supported
				UsageInKernelmode: data.CPU.SystemUsec * 1000,
				UsageInUsermode:   data.CPU.UserUsec * 1000,
			},
		}
	}

	if data.Memory != nil {
		st.MemoryStats = dockertypes.MemoryStats{
			Usage: data.Memory.Usage,
			// MaxUsage is not supported
			Limit: data.Memory.UsageLimit,
		}
		if data.MemoryEvents != nil {
			st.MemoryStats.Failcnt = data.MemoryEvents.Oom
		}
	}

	if data.Io != nil {
		var isbr []dockertypes.BlkioStatEntry
		for _, re := range data.Io.Usage {
			isbr = append(isbr,
				dockertypes.BlkioStatEntry{
					Major: re.Major,
					Minor: re.Minor,
					Op:    "read",
					Value: re.Rbytes,
				},
				dockertypes.BlkioStatEntry{
					Major: re.Major,
					Minor: re.Minor,
					Op:    "write",
					Value: re.Wbytes,
				},
			)
		}
		st.BlkioStats = dockertypes.BlkioStats{
			IoServiceBytesRecursive: isbr,
			// Other fields are unsupported
		}
	}

	return &st
}

func translateBlkioEntry(entries []*v1.BlkIOEntry) []dockertypes.BlkioStatEntry {
	out := make([]dockertypes.BlkioStatEntry, len(entries))
	for i, re := range entries {
		out[i] = dockertypes.BlkioStatEntry{
			Major: re.Major,
			Minor: re.Minor,
			Op:    re.Op,
			Value: re.Value,
		}
	}
	return out
}
