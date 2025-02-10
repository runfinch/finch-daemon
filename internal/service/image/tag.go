// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"fmt"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/idutil/imagewalker"
	"github.com/containerd/nerdctl/v2/pkg/referenceutil"

	eventtype "github.com/runfinch/finch-daemon/api/events"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

const tagEventAction = "tag"

func (s *service) Tag(ctx context.Context, srcImg string, repo, tag string) error {
	imgStore := s.client.GetClient().ImageService()
	srcImgName, err := s.getFullImageName(ctx, srcImg)
	if err != nil {
		return err
	}
	image, err := imgStore.Get(ctx, srcImgName)
	if err != nil {
		return err
	}
	rawRef := fmt.Sprintf("%s:%s", repo, tag)
	target, err := referenceutil.Parse(rawRef)
	if err != nil {
		return fmt.Errorf("target parse error: %w", err)
	}
	image.Name = target.String()
	if _, err = imgStore.Create(ctx, image); err != nil {
		if cerrdefs.IsAlreadyExists(err) {
			if err = imgStore.Delete(ctx, image.Name); err != nil {
				return err
			}
			if _, err = imgStore.Create(ctx, image); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	err = s.client.PublishEvent(ctx, tagTopic(), getTagEvent(image.Target.Digest.String(), rawRef))
	if err != nil {
		return err
	}

	return nil
}

func (s *service) getFullImageName(ctx context.Context, name string) (string, error) {
	var srcName string
	imgWalker := &imagewalker.ImageWalker{
		Client: s.client.GetClient(),
		OnFound: func(ctx context.Context, found imagewalker.Found) error {
			if srcName == "" {
				srcName = found.Image.Name
			}
			return nil
		},
	}
	matchCount, err := imgWalker.Walk(ctx, name)
	if err != nil {
		return "", fmt.Errorf("err from image walker: %w", err)
	}

	if matchCount < 1 {
		return "", errdefs.NewNotFound(fmt.Errorf("no such image: %s", name))
	}
	return srcName, nil
}

func tagTopic() string {
	return fmt.Sprintf("/%s/%s/%s", eventtype.CompatibleTopicPrefix, eventType, tagEventAction)
}

func getTagEvent(digest, imgName string) *eventtype.Event {
	return &eventtype.Event{
		ID:     digest,
		Status: tagEventAction,
		Type:   "image",
		Action: tagEventAction,
		Actor: eventtype.EventActor{
			Id: digest,
			Attributes: map[string]string{
				"name": imgName,
			},
		},
	}
}
