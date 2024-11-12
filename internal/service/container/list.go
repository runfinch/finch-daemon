// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"

	"github.com/runfinch/finch-daemon/api/types"
)

func (s *service) List(ctx context.Context, listOpts ncTypes.ContainerListOptions) ([]types.ContainerListItem, error) {
	ncContainers, err := s.nctlContainerSvc.ListContainers(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	containers := []types.ContainerListItem{}
	for _, ncc := range ncContainers {
		ncc.Names = fmt.Sprintf("/%s", ncc.Names)

		c, err := s.getContainer(ctx, ncc.ID)
		if err != nil {
			return nil, err
		}

		ci, err := s.nctlContainerSvc.InspectContainer(ctx, c)
		if err != nil {
			return nil, err
		}

		cli := types.ContainerListItem{
			Id:              ncc.ID,
			Names:           []string{ncc.Names},
			Image:           ncc.Image,
			CreatedAt:       ncc.CreatedAt.Unix(),
			State:           ci.State.Status,
			Labels:          ncc.Labels,
			NetworkSettings: ci.NetworkSettings,
			Mounts:          ci.Mounts,
		}

		l, err := c.Labels(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get container labels: %s", err)
		}
		updateNetworkSettings(ctx, cli.NetworkSettings, l)

		containers = append(containers, cli)
	}
	return containers, nil
}
