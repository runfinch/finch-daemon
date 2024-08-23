// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/containerd/containerd"
	"github.com/containerd/nerdctl/pkg/netutil"

	"github.com/runfinch/finch-daemon/api/handlers/network"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

var (
	errUnsupportedCNIDriver = errors.New("unsupported cni driver")
	errPluginNotFound       = errors.New("plugin not found")
	errNetworkIDNotFound    = errors.New("network ID not found")
)

type service struct {
	client    backend.ContainerdClient
	netClient backend.NerdctlNetworkSvc
	logger    flog.Logger
}

func NewService(client backend.ContainerdClient, netClient backend.NerdctlNetworkSvc, logger flog.Logger) network.Service {
	return &service{
		client:    client,
		netClient: netClient,
		logger:    logger,
	}
}

func (s *service) getNetwork(networkId string) (*netutil.NetworkConfig, error) {
	longIDExp, err := regexp.Compile(fmt.Sprintf("^sha256:%s.*", regexp.QuoteMeta(networkId)))
	if err != nil {
		return nil, err
	}

	shortIDExp, err := regexp.Compile(fmt.Sprintf("^%s", regexp.QuoteMeta(networkId)))
	if err != nil {
		return nil, err
	}

	idFilterFunc := func(n *netutil.NetworkConfig) bool {
		if n.NerdctlID == nil {
			// External network
			return n.Name == networkId
		}
		return n.Name == networkId || longIDExp.Match([]byte(*n.NerdctlID)) || shortIDExp.Match([]byte(*n.NerdctlID))
	}

	networks, err := s.netClient.FilterNetworks(idFilterFunc)
	if err != nil {
		s.logger.Errorf("failed to search network: %s. error: %s", networkId, err.Error())
		return nil, err
	}
	if len(networks) == 0 {
		s.logger.Debugf("no such network %s", networkId)
		return nil, errdefs.NewNotFound(fmt.Errorf("network %s not found", networkId))
	}
	if len(networks) > 1 {
		s.logger.Debugf("multiple IDs found with provided prefix: %s, total networks found: %d",
			networkId, len(networks))
		return nil, fmt.Errorf("multiple networks found with ID: %s", networkId)
	}

	return networks[0], nil
}

func (s *service) getContainer(ctx context.Context, containerId string) (containerd.Container, error) {
	searchResult, err := s.client.SearchContainer(ctx, containerId)
	if err != nil {
		s.logger.Errorf("failed to search container: %s. error: %s", containerId, err.Error())
		return nil, err
	}
	matchCount := len(searchResult)

	// if container not found then return NotFound error.
	if matchCount == 0 {
		s.logger.Debugf("no such container %s", containerId)
		return nil, errdefs.NewNotFound(fmt.Errorf("no such container %s", containerId))
	}
	// if more than one container found with the provided id return error.
	if matchCount > 1 {
		s.logger.Debugf("multiple IDs found with provided prefix: %s, total container found: %d",
			containerId, matchCount)
		return nil, fmt.Errorf("multiple IDs found with provided prefix: %s", containerId)
	}

	return searchResult[0], nil
}
