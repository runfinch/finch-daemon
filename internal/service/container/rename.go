// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Rename function renames a running container. It returns nil when it successfully renames the container.
func (s *service) Rename(ctx context.Context, cid string, newName string, opts ncTypes.ContainerRenameOptions) error {
	var err error
	con, _ := s.getContainer(ctx, newName)
	if con != nil {
		err = errdefs.NewConflict(fmt.Errorf("container with name %s already exists", newName))
		s.logger.Errorf("Failed to rename container: %s. Error: %v", cid, err)
		return err
	}

	con, err = s.getContainer(ctx, cid)
	if err != nil {
		return err
	}
	if err = s.nctlContainerSvc.RenameContainer(ctx, con, newName, opts); err != nil {
		s.logger.Errorf("Failed to rename container: %s. Error: %v", cid, err)
		return err
	}
	s.logger.Debugf("successfully renamed %s to %s", cid, newName)
	return nil
}
