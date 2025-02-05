// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"
	"time"

	containerd "github.com/containerd/containerd/v2/client"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Stop function stops a running container. It returns nil when it successfully stops the container.
func (s *service) Stop(ctx context.Context, cid string, timeout *time.Duration) error {
	con, err := s.getContainer(ctx, cid)
	if err != nil {
		return err
	}

	if s.isContainerStopped(ctx, con) {
		return errdefs.NewNotModified(fmt.Errorf("container is already stopped: %s", cid))
	}
	if err = s.nctlContainerSvc.StopContainer(ctx, con, timeout); err != nil {
		s.logger.Errorf("Failed to stop container: %s. Error: %v", cid, err)
		return err
	}
	s.logger.Debugf("successfully stopped: %s", cid)
	return nil
}

// isContainerStopped returns true when container is not in running state.
func (s *service) isContainerStopped(ctx context.Context, con containerd.Container) bool {
	status := s.client.GetContainerStatus(ctx, con)
	if status == containerd.Stopped || status == containerd.Created {
		return true
	}
	return false
}
