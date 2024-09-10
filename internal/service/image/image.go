// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/containerd/images"
	docker "github.com/distribution/reference"

	"github.com/runfinch/finch-daemon/api/handlers/image"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

// setting getAuthCredsFunc as a variable to allow mocking this function for unit testing.
var getAuthCredsFunc = (*service).getAuthCreds

type service struct {
	client       backend.ContainerdClient
	nctlImageSvc backend.NerdctlImageSvc
	logger       flog.Logger
}

func (s *service) getImage(ctx context.Context, name string) (*images.Image, error) {
	images, err := s.client.SearchImage(ctx, name)
	if err != nil {
		s.logger.Errorf("failed to search image: %s. error: %s", name, err)
		return nil, err
	}
	matchCount := len(images)

	// if image not found then return NotFound error
	if matchCount == 0 {
		s.logger.Debugf("no such image: %s", name)
		return nil, errdefs.NewNotFound(fmt.Errorf("no such image: %s", name))
	}

	// if multiple images are found, check that their digests are all the same, otherwise there could be a conflict
	if matchCount > 1 {
		var observedDigest string
		for _, img := range images {
			if observedDigest == "" {
				observedDigest = img.Target.Digest.String()
				continue
			}

			if observedDigest != img.Target.Digest.String() {
				s.logger.Debugf("multiple images with different digests found for %s", name)
				return nil, fmt.Errorf("multiple images with different digests found for %s", name)
			}
		}
	}

	// if one or more images are found with the same digest return the first one
	return &images[0], nil
}

func NewService(client backend.ContainerdClient, nerdctlImageSvc backend.NerdctlImageSvc, logger flog.Logger) image.Service {
	return &service{
		client:       client,
		nctlImageSvc: nerdctlImageSvc,
		logger:       logger,
	}
}

const (
	defaultTag      = "latest"
	tagDigestPrefix = "sha256:"
	eventType       = "image"
)

func canonicalize(name, tag string) (string, error) {
	if name != "" {
		if strings.HasPrefix(tag, tagDigestPrefix) {
			name += "@" + tag
		} else if tag != "" {
			name += ":" + tag
		}
	} else {
		name = tag
	}
	ref, err := docker.ParseAnyReference(name)
	if err != nil {
		return "", err
	}
	if named, ok := ref.(docker.Named); ok && refNeedsTag(ref) {
		tagged, err := docker.WithTag(named, defaultTag)
		if err == nil {
			ref = tagged
		}
	}
	return ref.String(), nil
}

func refNeedsTag(ref docker.Reference) bool {
	_, tagged := ref.(docker.Tagged)
	_, digested := ref.(docker.Digested)
	return !(tagged || digested)
}
