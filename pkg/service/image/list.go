// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"fmt"
	"log"

	"github.com/runfinch/finch-daemon/pkg/api/types"
)

func (s *service) List(ctx context.Context) ([]types.ImageSummary, error) {
	imgs, err := s.client.GetClient().ListImages(context.Background())
	if err != nil {
		return nil, fmt.Errorf("list: failed to list images: %w", err)
	}
	summaries := []types.ImageSummary{}
	// TODO post-process to dedup and populate RepoTags/RepoDigests
	// TODO walk image.Target() with the content store to get manifest digests?
	for _, img := range imgs {
		log.Println(img)
		if ok, err := img.IsUnpacked(ctx, "overlayfs"); !ok {
			continue
		} else if err != nil {
			return nil, err
		}
		info, err := img.ContentStore().Info(ctx, img.Target().Digest)
		if err != nil {
			return nil, err
		}
		size, err := img.Size(ctx)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, types.ImageSummary{
			ID: string(img.Target().Digest),
			RepoTags: []string{
				img.Name(),
			},
			Created: info.CreatedAt.Unix(),
			Size:    size,
		})
	}
	return summaries, nil
}
