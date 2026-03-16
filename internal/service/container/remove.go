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

	// PoC (Solution IV): clean up the pre-created netns and CNI network before
	// removing the container, so we don't leak network resources.
	lbls, labelErr := con.Labels(ctx)
	if labelErr == nil {
		netnsPath := lbls[labelNetnsPath]
		networkName := lbls[labelNetworkName]
		if netnsPath != "" && networkName != "" {
			if cleanupErr := s.nctlContainerSvc.RemoveContainerNetwork(ctx, con.ID(), networkName, netnsPath); cleanupErr != nil {
				// Log but don't block removal — a leaked netns is better than a
				// container that can never be removed.
				s.logger.Errorf("failed to cleanup container network on remove: %s", cleanupErr)
			}
		}
	} else {
		s.logger.Debugf("could not read container labels for network cleanup: %s", labelErr)
	}

	s.logger.Debugf("removing container: %s", con.ID())
	if err := s.nctlContainerSvc.RemoveContainer(ctx, con, force, removeVolumes); err != nil {
		if errors.As(err, &container.ErrContainerStatus{}) {
			s.logger.Debugf("Container is in running or pausing state. Failed to remove container: %s", con.ID())
			err = errdefs.NewConflict(fmt.Errorf("%s. unpause/stop container first or force removal", err))
			return err
		}
		s.logger.Errorf("Failed to remove container: %s. Error: %s", con.ID(), err.Error())
		return err
	}
	return nil
}
