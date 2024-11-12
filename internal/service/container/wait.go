// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"

	containerd "github.com/containerd/containerd/v2/client"
)

func (s *service) Wait(ctx context.Context, cid string, condition string) (code int64, err error) {
	con, err := s.getContainer(ctx, cid)
	// container wait status code is uint32, use -1 to indicate container search error
	if err != nil {
		return -1, err
	}
	s.logger.Debugf("wait container: %s", con.ID())
	rawcode, err := waitContainer(ctx, con)
	return int64(rawcode), err
}

// TODO: contribute to nerdctl to make this function public.
func waitContainer(ctx context.Context, container containerd.Container) (code uint32, err error) {
	task, err := container.Task(ctx, nil)
	if err != nil {
		return 0, err
	}

	statusC, err := task.Wait(ctx)
	if err != nil {
		return 0, err
	}

	status := <-statusC
	code, _, err = status.Result()

	return code, err
}
