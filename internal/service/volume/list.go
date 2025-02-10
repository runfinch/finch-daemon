// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"context"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"

	"github.com/runfinch/finch-daemon/api/types"
)

// List returns a list of volumes.
func (s *service) List(ctx context.Context, filters []string) (*types.VolumesListResponse, error) {
	// TODO: include size?
	vols, err := s.nctlVolumeSvc.ListVolumes(false, filters)
	if err != nil {
		s.logger.Errorf("failed to list volumes: %v", err)
		return nil, err
	}

	// initialize so empty response is [] instead of nil
	volumes := []native.Volume{}
	for _, vol := range vols {
		volumes = append(volumes, vol)
	}

	return &types.VolumesListResponse{Volumes: volumes}, nil
}
