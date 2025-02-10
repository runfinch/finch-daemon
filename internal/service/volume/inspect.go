// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"strings"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Inspect returns a details of a volume.
func (s *service) Inspect(name string) (*native.Volume, error) {
	vol, err := s.nctlVolumeSvc.GetVolume(name)
	if err != nil {
		// if the volume does not exist, return a NotFound error
		// see nerdctl code for exact error msg:
		// https://github.com/containerd/nerdctl/blob/main/pkg/mountutil/volumestore/volumestore.go#L134C3-L134C3
		if strings.Contains(err.Error(), "not found") {
			err = errdefs.NewNotFound(err)
		}
		return nil, err
	}
	return vol, err
}
