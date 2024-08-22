// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/containerd/containerd/images/converter"
	"github.com/containerd/nerdctl/pkg/imgutil/dockerconfigresolver"
	dockertypes "github.com/docker/cli/cli/config/types"

	"github.com/runfinch/finch-daemon/pkg/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Push(ctx context.Context, name, tag string, ac *dockertypes.AuthConfig, outStream io.Writer) (*types.PushResult, error) {
	// Canonicalize and parse raw image reference as "image:tag" or "image@digest"
	rawRef, err := canonicalize(name, tag)
	if err != nil {
		return nil, errdefs.NewInvalidFormat(fmt.Errorf("failed to canonicalize the ref: %w", err))
	}
	ref, refDomain, err := s.client.ParseDockerRef(rawRef)
	if err != nil {
		return nil, errdefs.NewInvalidFormat(err)
	}

	// Create a reduced platform image locally to avoid "400 Bad request" for multi-platform manifests
	// https://github.com/containerd/nerdctl/blob/v1.7.2/pkg/cmd/image/push.go#L93-L111
	platMC := s.client.DefaultPlatformStrict()
	pushRef := ref + "-tmp-reduced-platform"
	platImg, err := s.client.ConvertImage(ctx, pushRef, ref, converter.WithPlatform(platMC))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errdefs.NewNotFound(err)
		}
		return nil, fmt.Errorf("failed to create a tmp single-platform image %q: %s", pushRef, err)
	}
	if platImg != nil {
		defer s.client.DeleteImage(ctx, platImg.Name)
	}
	s.logger.Debugf("pushing as a reduced-platform image (%s, %s)", platImg.Target.MediaType, platImg.Target.Digest)

	// Get auth creds and the corresponding docker remotes resolver
	var creds dockerconfigresolver.AuthCreds
	if ac != nil {
		creds, err = getAuthCredsFunc(s, refDomain, *ac)
		if err != nil {
			return nil, err
		}
	}
	resolver, tracker, err := s.nctlImageSvc.GetDockerResolver(ctx, refDomain, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize remotes resolver: %s", err)
	}

	// finally, push the image
	if err = s.nctlImageSvc.PushImage(
		ctx,
		resolver,
		tracker,
		outStream,
		pushRef, ref,
		platMC,
	); err != nil {
		return nil, err
	}

	// send aux information for the pushed image
	img, err := s.client.GetImage(ctx, ref)
	if err != nil {
		return nil, nil
	}
	size, err := img.Size(ctx)
	if err != nil {
		return nil, nil
	}
	return &types.PushResult{
		Tag:    tag,
		Digest: img.Target().Digest.String(),
		Size:   int(size),
	}, nil
}
