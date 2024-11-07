// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package distribution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	dockerresolver "github.com/containerd/containerd/remotes/docker"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/pkg/imgutil/dockerconfigresolver"
	dockertypes "github.com/docker/cli/cli/config/types"
	registrytypes "github.com/docker/docker/api/types/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/runfinch/finch-daemon/api/handlers/distribution"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/runfinch/finch-daemon/pkg/flog"
	"github.com/runfinch/finch-daemon/pkg/utility/authutility"
	"github.com/runfinch/finch-daemon/pkg/utility/imageutility"
)

type service struct {
	client       backend.ContainerdClient
	nctlImageSvc backend.NerdctlImageSvc
	logger       flog.Logger
}

// setting getAuthCredsFunc as a variable to allow mocking this function for unit testing.
var getAuthCredsFunc = authutility.GetAuthCreds

func NewService(client backend.ContainerdClient, nerdctlImageSvc backend.NerdctlImageSvc, logger flog.Logger) distribution.Service {
	return &service{
		client:       client,
		nctlImageSvc: nerdctlImageSvc,
		logger:       logger,
	}
}

func (s *service) Inspect(ctx context.Context, name string, ac *dockertypes.AuthConfig) (*registrytypes.DistributionInspect, error) {
	// Canonicalize and parse raw image reference as "image:tag" or "image@digest"
	rawRef, err := imageutility.Canonicalize(name, "")
	if err != nil {
		return nil, errdefs.NewInvalidFormat(err)
	}
	namedRef, refDomain, err := s.client.ParseDockerRef(rawRef)
	if err != nil {
		return nil, errdefs.NewInvalidFormat(err)
	}

	// get auth creds and the corresponding docker remotes resolver
	var creds dockerconfigresolver.AuthCreds
	if ac != nil {
		creds, err = getAuthCredsFunc(refDomain, s.client, *ac)
		if err != nil {
			return nil, err
		}
	}
	resolver, _, err := s.nctlImageSvc.GetDockerResolver(ctx, refDomain, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize remotes resolver: %s", err)
	}

	_, desc, err := resolver.Resolve(ctx, namedRef)
	if err != nil {
		// translate error definitions from containerd
		switch {
		case cerrdefs.IsNotFound(err):
			return nil, errdefs.NewNotFound(err)
		case errors.Is(err, dockerresolver.ErrInvalidAuthorization):
			return nil, errdefs.NewForbidden(err)
		default:
			return nil, err
		}
	}

	fetcher, err := resolver.Fetcher(ctx, namedRef)
	if err != nil {
		return nil, err
	}

	rc, err := fetcher.Fetch(ctx, desc)
	if err != nil {
		switch {
		case cerrdefs.IsNotFound(err):
			return nil, errdefs.NewNotFound(err)
		default:
			return nil, err
		}
	}

	res, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	if dgst := desc.Digest.Algorithm().FromBytes(res); dgst != desc.Digest {
		return nil, fmt.Errorf("digest mismatch: %s != %s", dgst, desc.Digest)
	}

	var index ocispec.Index
	if err := json.Unmarshal(res, &index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index: %w", err)
	}

	var platforms []ocispec.Platform
	for _, manifest := range index.Manifests {
		platforms = append(platforms, *manifest.Platform)
	}

	return &registrytypes.DistributionInspect{
		Descriptor: ocispec.Descriptor{
			MediaType:    desc.MediaType,
			Digest:       desc.Digest,
			Size:         desc.Size,
			URLs:         desc.URLs,
			Annotations:  desc.Annotations,
			Data:         desc.Data,
			Platform:     desc.Platform,
			ArtifactType: desc.ArtifactType,
		},
		Platforms: platforms,
	}, nil
}
