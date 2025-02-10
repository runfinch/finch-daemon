// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"

	"github.com/containerd/nerdctl/v2/pkg/netutil"

	"github.com/runfinch/finch-daemon/api/types"
)

// List returns an array of network objects based on the filtered criteria. Nerdctl's pkg.cmd.network.List is not used
// as the output format is different from ours.
func (s *service) List(ctx context.Context) ([]*types.NetworkInspectResponse, error) {
	getAllFilterFunc := func(n *netutil.NetworkConfig) bool {
		return true
	}

	nl, err := s.netClient.FilterNetworks(getAllFilterFunc)
	if err != nil {
		return nil, err
	}

	summaries := make([]*types.NetworkInspectResponse, len(nl))

	for i, n := range nl {
		network := &types.NetworkInspectResponse{
			Name: n.Name,
		}
		if n.NerdctlID != nil {
			network.ID = *n.NerdctlID
		}
		summaries[i] = network
	}

	return summaries, nil
}
