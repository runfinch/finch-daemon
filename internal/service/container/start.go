// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	containerd "github.com/containerd/containerd/v2/client"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Start(ctx context.Context, cid string) error {
	cont, err := s.getContainer(ctx, cid)
	if err != nil {
		return err
	}
	if err := s.assertStartContainer(ctx, cont); err != nil {
		return err
	}
	// start the containers and if error occurs then return error otherwise return nil
	s.logger.Debugf("starting container: %s", cid)
	if err := s.nctlContainerSvc.StartContainer(ctx, cont); err != nil {
		s.logger.Errorf("Failed to start container: %s. Error: %v", cid, err)
		return err
	}
	s.logger.Debugf("successfully started: %s", cid)
	return nil
}

func (s *service) assertStartContainer(ctx context.Context, c containerd.Container) error {
	status := s.client.GetContainerStatus(ctx, c)
	switch status {
	case containerd.Running:
		return errdefs.NewNotModified(fmt.Errorf("container already running"))
	case containerd.Pausing:
	case containerd.Paused:
		return fmt.Errorf("cannot start a paused container, try unpause instead")
	}
	return nil
}
