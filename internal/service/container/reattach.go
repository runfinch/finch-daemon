// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/sirupsen/logrus"

	"github.com/runfinch/finch-daemon/internal/backend"
)

// ReattachRunningContainers resumes FIFO log copiers for containers that were
// running before the daemon restarted. Called once on daemon startup.
func ReattachRunningContainers(ctx context.Context, clientWrapper *backend.ContainerdClientWrapper, ncWrapper *backend.NerdctlWrapper) {
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

	// Try to reattach each container. Containers without running tasks are skipped.
	for _, c := range containers {
		reattachContainer(ctx, c, clientWrapper, dataStore)
	}
}

// reattachContainer reopens FIFOs for a single container and resumes log copiers.
// No-op if the container has no running task.
func reattachContainer(ctx context.Context, c containerd.Container, clientWrapper *backend.ContainerdClientWrapper, dataStore string) {
	containerLabels, err := c.Labels(ctx)
	if err != nil {
		return
	}
	ns := containerLabels[labels.Namespace]

	// Reattach to the existing task's FIFOs. containerd stores the FIFOSet
	// from task creation and returns it here, so we can reopen the read ends.
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
		return
	}

	// Only reattach copiers for containers that are actually running.
	status, err := task.Status(ctx)
	if err != nil || status.Status != containerd.Running {
		return
	}

	if dio == nil {
		return
	}

	// Resume log copiers using the same shared helper as customStart.
	logPath := containerLogPath(dataStore, ns, c.ID())
	startLogCopiers(logPath, dio)
	logrus.Infof("reattach: resumed log capture for container %s", c.ID())
}
