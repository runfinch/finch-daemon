// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Remove(ctx context.Context, networkId string) error {
	s.logger.Infof("network delete: network Id %s", networkId)
	net, err := s.getNetwork(networkId)
	if err != nil {
		return fmt.Errorf("failed to find network: %w", err)
	}
	usedNetworkInfo, err := s.netClient.UsedNetworkInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to find used network info: %w", err)
	}
	// API Doc: https://docs.docker.com/engine/api/v1.43/#tag/Network/operation/NetworkDelete
	// does not explicitly call out the scenario when network is in use by a container, although it returns a 403
	if value, ok := usedNetworkInfo[net.Name]; ok {
		return errdefs.NewForbidden(fmt.Errorf("network %q is in use by container %q", networkId, value))
	}
	if net.File == "" {
		return errdefs.NewForbidden(fmt.Errorf("%s is a pre-defined network and cannot be removed", networkId))
	}
	return s.netClient.RemoveNetwork(net)
}
