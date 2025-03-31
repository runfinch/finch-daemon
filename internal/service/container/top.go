// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"

	cerrdefs "github.com/containerd/errdefs"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Top(ctx context.Context, cid string, options ncTypes.ContainerTopOptions) error {
	_, err := s.getContainer(ctx, cid)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return errdefs.NewNotFound(err)
		}
		return err
	}

	err = s.nctlContainerSvc.ContainerTop(ctx, cid, options)
	if err != nil {
		return err
	}

	return nil
}
