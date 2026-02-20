// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"io"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func (s *service) Export(ctx context.Context, name string, platform *ocispec.Platform, outStream io.Writer) error {
	img, err := s.getImage(ctx, name)
	if err != nil {
		return err
	}
	return s.nctlImageSvc.ExportImage(ctx, []string{img.Name}, platform, outStream)
}
