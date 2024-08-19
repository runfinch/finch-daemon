// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"fmt"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Remove(ctx context.Context, name string, force bool) (deleted, untagged []string, err error) {
	matchCount, uniqueCount, imgs, err := s.nctlImageSvc.SearchImage(ctx, name)
	if err != nil {
		return
	}
	if matchCount == 0 {
		err = errdefs.NewNotFound(fmt.Errorf("no such image: %s", name))
		return
	}
	if matchCount > 1 && !(force && uniqueCount == 1) {
		err = errdefs.NewConflict(fmt.Errorf(
			"unable to delete %s (must be forced) - image is referenced in multiple repositories", name))
		return
	}
	// check if the image can be deleted
	stoppedImgs, runningImgs, err := s.client.GetUsedImages(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, img := range imgs {
		if cid, ok := runningImgs[img.Name]; ok {
			err = fmt.Errorf("unable to delete %s (cannot be forced) - image is being used by running container %s", name, cid)
			return nil, nil, errdefs.NewConflict(err)
		}
		if cid, ok := stoppedImgs[img.Name]; ok && !force {
			err = fmt.Errorf("unable to delete %s (must be forced) - image is being used by stopped container %s", name, cid)
			return nil, nil, errdefs.NewConflict(err)
		}
	}

	//delete image
	deleted = []string{}
	untagged = []string{}
	for _, img := range imgs {
		digests, err := s.client.GetImageDigests(ctx, img)
		if err != nil {
			s.logger.Warnf("Failed to enumerate rootfs. Error: %s", err)
		}
		err = s.client.DeleteImage(ctx, img.Name)
		if err != nil {
			return nil, nil, err
		}

		// TODO: a digest only gets deleted when all the images ref is deleted. Need to fix this later.
		// Nerdctl also has the same problem. it also reports digest got deleted even there are other images
		// reference to that digest.
		for _, d := range digests {
			deleted = append(deleted, d.String())
		}
		untagged = append(untagged, fmt.Sprintf("%s:%s", img.Name, img.Target.Digest))
	}
	return untagged, deleted, err
}
