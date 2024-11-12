// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/containerd/containerd/remotes/docker"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/imgutil/dockerconfigresolver"
	dockertypes "github.com/docker/cli/cli/config/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Pull(ctx context.Context, name, tag, platformStr string, ac *dockertypes.AuthConfig, outStream io.Writer) error {
	// get host platform's default spec if unspecified
	var platform ocispec.Platform
	var err error
	if platformStr == "" {
		platform = s.client.DefaultPlatformSpec()
	} else {
		platform, err = s.client.ParsePlatform(platformStr)
	}
	if err != nil {
		return fmt.Errorf("invalid platform %s: %s", platformStr, err)
	}

	// parse image reference into registry hostname and image name
	rawRef := toImageRef(name, tag)
	ref, refDomain, err := s.client.ParseDockerRef(rawRef)
	if err != nil {
		return errdefs.NewInvalidFormat(err)
	}

	// get auth creds and the corresponding docker remotes resolver
	var creds dockerconfigresolver.AuthCreds
	if ac != nil {
		creds, err = getAuthCredsFunc(refDomain, s.client, *ac)
		if err != nil {
			return err
		}
	}
	resolver, _, err := s.nctlImageSvc.GetDockerResolver(ctx, refDomain, creds)
	if err != nil {
		return fmt.Errorf("failed to initialize remotes resolver: %s", err)
	}

	// finally, pull the image
	_, err = s.nctlImageSvc.PullImage(
		ctx,
		outStream, outStream,
		resolver,
		ref,
		[]ocispec.Platform{platform},
	)

	if err != nil {
		if errors.Is(err, docker.ErrInvalidAuthorization) || cerrdefs.IsNotFound(err) {
			err = errdefs.NewNotFound(err)
		}

		// nerdctl issue: if there is an error during pull, it is returned before the
		// progress writer is shutdown properly. This can cause panic as the progress writer
		// tries to write to the http stream writer, which is nil after the handler returns.
		// Wait 100ms to give progress writer enough time to exit.
		//
		// TODO: Fix upstream. https://github.com/containerd/nerdctl/blob/v1.4.0/pkg/imgutil/pull/pull.go#L95-L101
		time.Sleep(time.Millisecond * 100)
	}
	return err
}

func toImageRef(name, tag string) string {
	if tag == "" {
		return name
	}
	// Handle the case where the tag starts with a digest algorithm. We do not
	// handle digests specified without an algorithm.
	if strings.HasPrefix(tag, "sha256:") {
		return fmt.Sprintf("%s@%s", name, tag)
	}
	return fmt.Sprintf("%s:%s", name, tag)
}
