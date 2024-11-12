// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package container comprises functions and structures related to container APIs
package container

import (
	"context"
	"fmt"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/spf13/afero"

	"github.com/runfinch/finch-daemon/api/handlers/container"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/archive"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/runfinch/finch-daemon/pkg/flog"
	"github.com/runfinch/finch-daemon/pkg/statsutil"
)

type NerdctlService interface {
	backend.NerdctlContainerSvc
	backend.NerdctlNetworkSvc
}

type service struct {
	client           backend.ContainerdClient
	nctlContainerSvc NerdctlService
	logger           flog.Logger
	fs               afero.Fs
	tarCreator       archive.TarCreator
	tarExtractor     archive.TarExtractor
	stats            statsutil.StatsUtil
}

// NewService creates a new service to operate on containers.
func NewService(
	client backend.ContainerdClient,
	nerdctlContainerSvc NerdctlService,
	logger flog.Logger,
	fs afero.Fs,
	tarCreator archive.TarCreator,
	tarExtractor archive.TarExtractor,
) container.Service {
	return &service{
		client:           client,
		nctlContainerSvc: nerdctlContainerSvc,
		logger:           logger,
		fs:               fs,
		tarCreator:       tarCreator,
		tarExtractor:     tarExtractor,
		stats:            statsutil.NewStatsUtil(),
	}
}

// getContainer returns a containerd container from container id.
func (s *service) getContainer(ctx context.Context, cid string) (containerd.Container, error) {
	searchResult, err := s.client.SearchContainer(ctx, cid)
	if err != nil {
		s.logger.Errorf("failed to search container: %s. error: %s", cid, err.Error())
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
