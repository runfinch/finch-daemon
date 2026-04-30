// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/containerd/nerdctl/v2/pkg/ocihook"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"

	"github.com/runfinch/finch-daemon/internal/backend"
)

// ReattachRunningContainers resumes FIFO log copiers and postStop watchers for
// containers that were running before the daemon restarted, and cleans up
// orphaned CNI state for containers that stopped while the daemon was down.
// Called once on daemon startup from main.go.
func ReattachRunningContainers(ctx context.Context, clientWrapper *backend.ContainerdClientWrapper, ncWrapper *backend.NerdctlWrapper) {
	// List all containers known to containerd.
	containers, err := clientWrapper.GetContainers(ctx)
	if err != nil {
		logrus.Warnf("reattach: failed to list containers: %v", err)
		return
	}

	dataStore, err := ncWrapper.GetDataStore()
	if err != nil {
		logrus.Warnf("reattach: failed to get data store: %v", err)
		return
	}
	globalOpts := ncWrapper.GetGlobalOptions()
	runningIDs := make(map[string]bool)

	// Reattach each container. Non-running containers are skipped.
	for _, c := range containers {
		if reattachContainer(ctx, c, clientWrapper, dataStore, globalOpts) {
			runningIDs[c.ID()] = true
		}
	}

	// Clean up CNI state for containers that stopped while the daemon was down.
	// These containers have no postStop watcher to clean up after them.
	cleanupOrphanedCNIState(ctx, containers, runningIDs, dataStore, globalOpts)
}

// reattachContainer reopens FIFOs for a single container, resumes log copiers,
// and starts a postStop watcher for CNI teardown. Returns true if the container
// has a running task.
func reattachContainer(ctx context.Context, c containerd.Container, clientWrapper *backend.ContainerdClientWrapper, dataStore string, globalOpts *types.GlobalCommandOptions) bool {
	// Get namespace label — needed for log path and CNI bundle path.
	containerLabels, err := c.Labels(ctx)
	if err != nil {
		return false
	}
	ns := containerLabels[labels.Namespace]

	// Only reattach containers started via the HTTP API (customStart).
	// Containers started via nerdctl CLI use binary-based logging and OCI hooks;
	// reopening their FIFOs or attaching postStop watchers would interfere.
	// customStart-created tasks use FIFO-based IO, which is reflected by the
	// absence of a binary:// LogURI or by having no LogURI at all after hook stripping.
	logURI := containerLabels[labels.LogURI]
	if strings.HasPrefix(logURI, "binary://") {
		return false
	}

	// Reattach to the existing task's FIFOs. containerd stores the FIFOSet
	// from task creation and returns it here so we can reopen the read ends.
	var dio *cio.DirectIO
	attachFunc := func(fifos *cio.FIFOSet) (cio.IO, error) {
		fifos.Stdin = ""
		dio, err = clientWrapper.NewDirectCIO(ctx, fifos)
		if err != nil {
			return nil, err
		}
		return dio, nil
	}

	// Task() with an attach function returns the existing task and reopens IO.
	// Returns error if no task exists (container is stopped).
	task, err := c.Task(ctx, attachFunc)
	if err != nil {
		return false
	}

	// Only reattach for containers that are actually running.
	status, err := task.Status(ctx)
	if err != nil || status.Status != containerd.Running {
		return false
	}

	// attachFunc may not have been called if IO wasn't set up.
	if dio == nil {
		return true
	}

	// Resume log copiers using the same shared helper as customStart.
	logPath := containerLogPath(dataStore, ns, c.ID())
	startLogCopiers(logPath, dio)
	logrus.Infof("reattach: resumed log capture for container %s", c.ID())

	// Resume postStop watcher — runs CNI teardown when container exits.
	// Without this, containers that exit after daemon restart would leak CNI state.
	watchPostStop(task, c.ID(), containerLabels, ns, dataStore, globalOpts)
	logrus.Infof("reattach: resumed postStop watcher for container %s", c.ID())

	return true
}

// runPostStop calls ocihook.Run("postStop") to tear down CNI networking.
// Shared by watchPostStop (normal exit path) and cleanupOrphanedCNIState
// (daemon restart recovery path).
func runPostStop(containerID string, containerLabels map[string]string, ns, dataStore string, globalOpts *types.GlobalCommandOptions) {
	// Skip CNI teardown if CNI paths are not configured.
	if globalOpts.CNIPath == "" || globalOpts.CNINetConfPath == "" {
		return
	}

	// Construct the OCI state that ocihook.Run expects on stdin.
	// Status is Stopped, Pid is 0 (process no longer exists).
	state := specs.State{
		Version:     "1.0.0",
		ID:          containerID,
		Status:      specs.StateStopped,
		Pid:         0,
		Bundle:      fmt.Sprintf("/run/containerd/io.containerd.runtime.v2.task/%s/%s", ns, containerID),
		Annotations: containerLabels,
	}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		logrus.Warnf("postStop: failed to marshal state for container %s: %v", containerID, err)
		return
	}

	// ocihook.Run("postStop") calls cni.Remove to release IP, clean iptables,
	// and remove hostsstore/namestore entries.
	if err := ocihook.Run(
		bytes.NewReader(stateJSON),
		os.Stderr,
		"postStop",
		dataStore,
		globalOpts.CNIPath,
		globalOpts.CNINetConfPath,
		globalOpts.BridgeIP,
	); err != nil {
		logrus.Warnf("postStop: CNI teardown failed for container %s: %v", containerID, err)
	} else {
		logrus.Infof("postStop: CNI teardown complete for container %s", containerID)
	}
}

// cleanupOrphanedCNIState scans /var/lib/cni/results/ for CNI allocations
// belonging to containers that are no longer running. This handles the case
// where the daemon crashed while a container was running, and the container
// subsequently stopped before the daemon restarted — no postStop watcher
// was alive to clean up.
func cleanupOrphanedCNIState(ctx context.Context, containers []containerd.Container, runningIDs map[string]bool, dataStore string, globalOpts *types.GlobalCommandOptions) {
	// CNI result files are stored here by the CNI plugins.
	cniResultsDir := "/var/lib/cni/results"
	entries, err := os.ReadDir(cniResultsDir)
	if err != nil {
		return // directory may not exist — no CNI state to clean
	}

	// Build a lookup map of all known containers.
	allContainers := make(map[string]containerd.Container)
	for _, c := range containers {
		allContainers[c.ID()] = c
	}

	// For each CNI result file, check if it belongs to a non-running container.
	// CNI result filenames contain the container ID (e.g., bridge-finch-{id}-eth0).
	// Only clean up containers that were in our snapshot and confirmed not running.
	// Skip containers not in our snapshot — they may have been created after we
	// listed containers and could still be starting up.
	cleaned := 0
	for _, entry := range entries {
		name := entry.Name()
		for containerID := range allContainers {
			if strings.Contains(name, containerID) && !runningIDs[containerID] {
				c := allContainers[containerID]
				containerLabels, err := c.Labels(ctx)
				if err != nil {
					continue
				}
				// Skip containers created by nerdctl CLI — their CNI state
				// is managed by nerdctl's own hooks, not by finch-daemon.
				logURI := containerLabels[labels.LogURI]
				if strings.HasPrefix(logURI, "binary://") {
					break
				}
				ns := containerLabels[labels.Namespace]
				logrus.Infof("reattach: cleaning up orphaned CNI state for container %s", containerID)
				runPostStop(containerID, containerLabels, ns, dataStore, globalOpts)
				cleaned++
				break
			}
		}
	}

	if cleaned > 0 {
		logrus.Infof("reattach: cleaned up orphaned CNI state for %d containers", cleaned)
	}
}
