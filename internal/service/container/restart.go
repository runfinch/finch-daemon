// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"time"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Restart(ctx context.Context, cid string, timeout time.Duration) error {
	con, err := s.getContainer(ctx, cid)
	if err != nil {
		return err
	}

	// restart the container and if error occurs then return error otherwise return nil
	// swallow IsNotModified error on StopContainer for already stopped container, simply call StartContainer
	s.logger.Debugf("restarting container: %s", cid)
	if err := s.nctlContainerSvc.StopContainer(ctx, con, &timeout); err != nil && !errdefs.IsNotModified(err) {
		s.logger.Errorf("Failed to stop container: %s. Error: %v", cid, err)
		return err
	}
	if err = s.nctlContainerSvc.StartContainer(ctx, con); err != nil {
		s.logger.Errorf("Failed to start container: %s. Error: %v", cid, err)
		return err
	}
	s.logger.Debugf("successfully restarted: %s", cid)
	return nil
}
