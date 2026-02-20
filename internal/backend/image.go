// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"
	"io"

	"github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/nerdctl/v2/pkg/cmd/image"
	"github.com/containerd/containerd/v2/core/remotes"
	"github.com/containerd/containerd/v2/core/remotes/docker"
	dockerconfig "github.com/containerd/containerd/v2/core/remotes/docker/config"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/containerdutil"
	"github.com/containerd/nerdctl/v2/pkg/idutil/imagewalker"
	"github.com/containerd/nerdctl/v2/pkg/imageinspector"
	"github.com/containerd/nerdctl/v2/pkg/imgutil"
	"github.com/containerd/nerdctl/v2/pkg/imgutil/dockerconfigresolver"
	"github.com/containerd/nerdctl/v2/pkg/imgutil/load"
	"github.com/containerd/nerdctl/v2/pkg/imgutil/push"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/containerd/platforms"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

//go:generate mockgen --destination=../../mocks/mocks_backend/nerdctlimagesvc.go -package=mocks_backend github.com/runfinch/finch-daemon/internal/backend NerdctlImageSvc
type NerdctlImageSvc interface {
	InspectImage(ctx context.Context, image images.Image) (*dockercompat.Image, error)
	GetDockerResolver(ctx context.Context, refDomain string, creds dockerconfigresolver.AuthCreds) (remotes.Resolver, docker.StatusTracker, error)
	PullImage(ctx context.Context, stdout, stderr io.Writer, resolver remotes.Resolver, ref string, platforms []ocispec.Platform) (*imgutil.EnsuredImage, error)
	PushImage(ctx context.Context, resolver remotes.Resolver, tracker docker.StatusTracker, stdout io.Writer, pushRef, ref string, platMC platforms.MatchComparer) error
	SearchImage(ctx context.Context, name string) (int, int, []*images.Image, error)
	LoadImage(ctx context.Context, img string, stdout io.Writer, quiet bool) error
	ExportImage(ctx context.Context, imageNames []string, platform *ocispec.Platform, writer io.Writer) error
	GetDataStore() (string, error)
	Namespace() string
}

func (w *NerdctlWrapper) InspectImage(ctx context.Context, image images.Image) (*dockercompat.Image, error) {
	snapshotter := containerdutil.SnapshotService(w.clientWrapper.client, w.globalOptions.Snapshotter)
	n, err := imageinspector.Inspect(ctx, w.clientWrapper.client, image, snapshotter)
	if err != nil {
		return nil, err
	}
	return dockercompat.ImageFromNative(n)
}

// GetDockerResolver returns a new Docker config resolver from the reference host and auth credentials.
func (w *NerdctlWrapper) GetDockerResolver(ctx context.Context, refDomain string, creds dockerconfigresolver.AuthCreds) (remotes.Resolver, docker.StatusTracker, error) {
	dOpts := []dockerconfigresolver.Opt{dockerconfigresolver.WithHostsDirs(w.globalOptions.HostsDir)}
	if creds != nil {
		dOpts = append(dOpts, dockerconfigresolver.WithAuthCreds(creds))
	}

	hostOpts, err := dockerconfigresolver.NewHostOptions(ctx, refDomain, dOpts...)
	if err != nil {
		return nil, nil, err
	}

	tracker := docker.NewInMemoryTracker()
	resolverOpts := docker.ResolverOptions{
		Tracker: tracker,
		Hosts:   dockerconfig.ConfigureHosts(ctx, *hostOpts),
	}

	return docker.NewResolver(resolverOpts), tracker, nil
}

// PullImage pulls an image from nerdctl's imgutil library.
func (w *NerdctlWrapper) PullImage(ctx context.Context, stdout, stderr io.Writer, resolver remotes.Resolver, ref string, platforms []ocispec.Platform) (*imgutil.EnsuredImage, error) {
	opts := types.ImagePullOptions{
		Stdout:          stdout,
		Stderr:          stderr,
		GOptions:        *w.globalOptions,
		Unpack:          nil,
		OCISpecPlatform: platforms,
		Mode:            "always",
		Quiet:           false,
	}

	return imgutil.PullImage(
		ctx,
		w.clientWrapper.client,
		resolver,
		ref,
		opts,
	)
}

// PushImage pushes an image using nerdctl's imgutil library.
func (w *NerdctlWrapper) PushImage(ctx context.Context, resolver remotes.Resolver, tracker docker.StatusTracker, stdout io.Writer, pushRef, ref string, platMC platforms.MatchComparer) error {
	return push.Push(
		ctx,
		w.clientWrapper.client,
		resolver,
		tracker,
		stdout,
		pushRef, ref,
		platMC,
		false,
		false,
	)
}

func (w *NerdctlWrapper) SearchImage(ctx context.Context, name string) (int, int, []*images.Image, error) {
	uniqueCount := 0
	var imgs []*images.Image
	walker := &imagewalker.ImageWalker{
		Client: w.clientWrapper.GetClient(),
		OnFound: func(ctx context.Context, found imagewalker.Found) error {
			uniqueCount = found.UniqueImages
			imgs = append(imgs, &found.Image)
			return nil
		},
	}
	n, err := walker.Walk(ctx, name)
	return n, uniqueCount, imgs, err
}

func (w *NerdctlWrapper) LoadImage(ctx context.Context, img string, stdout io.Writer, q bool) error {
	// TODO currently the "quiet" flag in nerdctl is hardcoded as "false".
	// Ideally this flag should be part of the ImageLoadOptions, we can
	// contribute this enhancement at upstream
	_, err := load.FromArchive(ctx, w.clientWrapper.client, types.ImageLoadOptions{
		Stdout:       stdout,
		GOptions:     *w.globalOptions,
		Input:        img,
		AllPlatforms: true,
		Quiet:        q,
	})
	return err
}

func (w *NerdctlWrapper) ExportImage(ctx context.Context, imageNames []string, platform *ocispec.Platform, writer io.Writer) error {
	opts := types.ImageSaveOptions{
		Stdout:   writer,
		GOptions: *w.globalOptions,
	}
	if platform != nil {
		opts.Platform = []string{platforms.Format(*platform)}
	} else {
		opts.AllPlatforms = true
	}
	return image.Save(ctx, w.clientWrapper.client, imageNames, opts)
}
