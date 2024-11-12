// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"
	"fmt"

	"github.com/containerd/nerdctl/v2/pkg/cmd/container"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Remove function deletes a container. It returns nil when it successfully removes the container.
func (s *service) Remove(ctx context.Context, cid string, force, removeVolumes bool) (err error) {
	con, err := s.getContainer(ctx, cid)
	if err != nil {
		return err
	}

	s.logger.Debugf("removing container: %s", con.ID())
	if err := s.nctlContainerSvc.RemoveContainer(ctx, con, force, removeVolumes); err != nil {
		// if containers is running then return with proper error msg otherwise return the original error msg
		if errors.As(err, &container.ErrContainerStatus{}) {
			s.logger.Debugf("Container is in running or pausing state. Failed to remove container: %s", con.ID())
			err = errdefs.NewConflict(fmt.Errorf("%s. unpause/stop container first or force removal", err))
			return err
		}
		// failed to delete the container. log the error msg and return with error
		s.logger.Errorf("Failed to remove container: %s. Error: %s", con.ID(), err.Error())
		return err
	}
	return nil
}
