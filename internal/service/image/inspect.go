// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
)

func (s *service) Inspect(ctx context.Context, name string) (*dockercompat.Image, error) {
	img, err := s.getImage(ctx, name)
	if err != nil {
		return nil, err
	}

	image, err := s.nctlImageSvc.InspectImage(ctx, *img)
	if err != nil {
		return nil, err
	}

	// reset image Id so that it matches image digest (nerdctl compatible) instead of docker-compatible id
	image.ID = img.Target.Digest.String()
	return image, nil
}
