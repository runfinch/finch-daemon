// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"bufio"
	"bytes"
	"context"
	"strings"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Remove Delete a volume from the system.
func (s *service) Remove(ctx context.Context, volName string, force bool) error {
	// pass a dummy writer to the nerdctl, since the stdout output is not required for the remove operation
	var buf bytes.Buffer
	dummyWriter := bufio.NewWriter(&buf)

	err := s.nctlVolumeSvc.RemoveVolume(ctx, volName, force, dummyWriter)
	if err != nil {
		// convert the nerdctl error to a finch specific error to return the appropriate status code
		switch {
		case strings.Contains(err.Error(), "not found"):
			err = errdefs.NewNotFound(err)
		case strings.Contains(err.Error(), "in use"):
			err = errdefs.NewConflict(err)
		case strings.Contains(err.Error(), "could not be removed"):
			err = errdefs.NewInvalidFormat(err)
		}
		return err
	}
	return nil
}
