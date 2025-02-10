// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"fmt"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	cerrdefs "github.com/containerd/errdefs"

	"github.com/runfinch/finch-daemon/api/handlers/exec"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

type service struct {
	client backend.ContainerdClient
	logger flog.Logger
}

// NewService creates a new service to run exec processes.
func NewService(
	client backend.ContainerdClient,
	logger flog.Logger,
) exec.Service {
	return &service{
		client: client,
		logger: logger,
	}
}

type execInstance struct {
	Container containerd.Container
	Task      containerd.Task
	Process   containerd.Process
}

func (s *service) getContainer(ctx context.Context, cid string) (containerd.Container, error) {
	searchResult, err := s.client.SearchContainer(ctx, cid)
	if err != nil {
		s.logger.Errorf("failed to search container: %s. error: %v", cid, err)
		return nil, err
	}
	matchCount := len(searchResult)

	// if container not found then return NotFound error.
	if matchCount == 0 {
		s.logger.Debugf("no such container: %s", cid)
		return nil, errdefs.NewNotFound(fmt.Errorf("no such container: %s", cid))
	}
	// if more than one container found with the provided id return error.
	if matchCount > 1 {
		s.logger.Debugf("multiple IDs found with provided prefix: %s, total containers found: %d", cid, matchCount)
		return nil, fmt.Errorf("multiple IDs found with provided prefix: %s", cid)
	}

	return searchResult[0], nil
}

func (s *service) loadExecInstance(ctx context.Context, conID, execID string, attach cio.Attach) (*execInstance, error) {
	con, err := s.getContainer(ctx, conID)
	if err != nil {
		if cerrdefs.IsNotFound(err) || errdefs.IsNotFound(err) {
			return nil, errdefs.NewNotFound(fmt.Errorf("container not found: %v", err))
		}
		return nil, err
	}

	task, err := con.Task(ctx, nil)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return nil, errdefs.NewNotFound(fmt.Errorf("task not found: %v", err))
		}
		return nil, err
	}

	proc, err := task.LoadProcess(ctx, execID, attach)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return nil, errdefs.NewNotFound(fmt.Errorf("process not found: %v", err))
		}
		return nil, err
	}

	return &execInstance{
		Container: con,
		Task:      task,
		Process:   proc,
	}, nil
}
