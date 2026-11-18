// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/defaults"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/labels"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Start(ctx context.Context, cid string, options types.ContainerStartOptions) error {
	cont, err := s.getContainer(ctx, cid)
	if err != nil {
		return err
	}
	if err := s.assertStartContainer(ctx, cont); err != nil {
		return err
	}
	s.logger.Debugf("starting container: %s", cid)

	if err := s.customStart(ctx, cid, cont, options); err != nil {
		s.logger.Errorf("Failed to start container: %s. Error: %v", cid, err)
		return err
	}

	s.logger.Debugf("successfully started: %s", cid)
	return nil
}

// customStart creates a container task with FIFO-based logging and starts it.
// It replaces nerdctl's container.Start() to give finch-daemon direct control
// over IO and task lifecycle.
func (s *service) customStart(ctx context.Context, cid string, cont containerd.Container, options types.ContainerStartOptions) error {
	s.cleanupOldTask(ctx, cont)

	// Create a new task with FIFO-based IO for log capture.
	task, directIO, err := s.createTaskWithFIFO(ctx, cont)
	if err != nil {
		return err
	}

	containerLabels, err := cont.Labels(ctx)
	if err != nil {
		task.Delete(ctx)
		return fmt.Errorf("failed to get container labels: %w", err)
	}
	ns := containerLabels[labels.Namespace]

	dataStore, err := s.nctlContainerSvc.GetDataStore()
	if err != nil {
		task.Delete(ctx)
		return fmt.Errorf("failed to get data store: %w", err)
	}

	logPath := containerLogPath(dataStore, ns, cont.ID())

	if err := task.Start(ctx); err != nil {
		return err
	}

	// Start copiers after task.Start to avoid goroutine leaks if Start fails.
	startLogCopiers(logPath, directIO)

	return nil
}

// cleanupOldTask deletes any existing task for the container. This handles
// the case where a previous task exists from a prior run (e.g., container restart).
func (s *service) cleanupOldTask(ctx context.Context, cont containerd.Container) {
	if oldTask, err := cont.Task(ctx, nil); err == nil {
		if _, err := oldTask.Delete(ctx); err != nil {
			s.logger.Debugf("failed to delete old task for %s: %v", cont.ID(), err)
		}
	}
}

// createTaskWithFIFO creates a new container task using FIFO-based IO instead of
// the binary:// log driver. containerd's shim writes container stdout/stderr to
// the FIFO write ends; the returned DirectIO provides the read ends that our
// copier goroutines (see copier.go) consume.
func (s *service) createTaskWithFIFO(ctx context.Context, cont containerd.Container) (containerd.Task, *cio.DirectIO, error) {
	spec, err := cont.Spec(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get container spec: %w", err)
	}

	// Terminal mode multiplexes stdout+stderr into a single stream.
	isTerminal := spec.Process != nil && spec.Process.Terminal

	var directIO *cio.DirectIO
	ioCreator := func(id string) (cio.IO, error) {
		// Create named pipes under containerd's default FIFO directory.
		fifos, err := s.client.NewFIFOSetInDir(defaults.DefaultFIFODir, id, isTerminal)
		if err != nil {
			return nil, err
		}
		fifos.Stdin = ""

		// Open the FIFOs and return DirectIO with read ends for stdout/stderr.
		directIO, err = s.client.NewDirectCIO(ctx, fifos)
		if err != nil {
			return nil, err
		}
		return directIO, nil
	}

	task, err := cont.NewTask(ctx, ioCreator)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, directIO, nil
}

func (s *service) assertStartContainer(ctx context.Context, c containerd.Container) error {
	status := s.client.GetContainerStatus(ctx, c)
	switch status {
	case containerd.Running:
		return errdefs.NewNotModified(fmt.Errorf("container already running"))
	case containerd.Pausing:
	case containerd.Paused:
		return fmt.Errorf("cannot start a paused container, try unpause instead")
	}
	return nil
}
