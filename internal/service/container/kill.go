// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	containerd "github.com/containerd/containerd/v2/client"
	cerrdefs "github.com/containerd/errdefs"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Kill(ctx context.Context, cid string, options ncTypes.ContainerKillOptions) error {
	cont, err := s.getContainer(ctx, cid)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return errdefs.NewNotFound(err)
		}
		return err
	}
	status := s.client.GetContainerStatus(ctx, cont)
	if status != containerd.Running {
		return errdefs.NewConflict(fmt.Errorf("container %s is not running", cid))
	}

	err = s.nctlContainerSvc.KillContainer(ctx, cid, options)
	if err != nil {
		return err
	}

	return nil
}
