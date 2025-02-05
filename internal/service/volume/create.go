// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"context"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
)

// Create a new volume and return the pointer to that volume.
func (s *service) Create(ctx context.Context, name string, labels []string) (*native.Volume, error) {
	newVolume, err := s.nctlVolumeSvc.CreateVolume(name, labels)
	if err != nil {
		s.logger.Errorf("failed to create volume: %v", err)
		return nil, err
	}

	return newVolume, nil
}
