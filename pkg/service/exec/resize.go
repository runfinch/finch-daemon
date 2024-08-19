// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"

	"github.com/runfinch/finch-daemon/pkg/api/types"
)

func (s *service) Resize(ctx context.Context, options *types.ExecResizeOptions) error {
	exec, err := s.loadExecInstance(ctx, options.ConID, options.ExecID, nil)
	if err != nil {
		return err
	}

	return exec.Process.Resize(ctx, uint32(options.Width), uint32(options.Height))
}
