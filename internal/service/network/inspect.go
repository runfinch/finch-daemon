// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"

	"github.com/runfinch/finch-daemon/api/types"
)

// Inspect returns the Name and Id of the network given an Id or the Name.
func (s *service) Inspect(ctx context.Context, networkId string) (*types.NetworkInspectResponse, error) {
	s.logger.Infof("network inspect: network Id %s", networkId)
	n, err := s.getNetwork(networkId)
	if err != nil {
		s.logger.Debugf("Failed to get network: %s", err)
		return nil, err
	}
	network, err := s.netClient.InspectNetwork(ctx, n)
	if err != nil {
		s.logger.Debugf("Failed to inspect network: %s", err)
		return nil, err
	}

	netObject := &types.NetworkInspectResponse{
		Name:   network.Name,
		ID:     network.ID,
		IPAM:   network.IPAM,
		Labels: network.Labels,
	}

	return netObject, nil
}
