// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"
	"fmt"
	"regexp"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/events"
	"github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/core/images/converter"
	"github.com/containerd/containerd/v2/core/mount"
	"github.com/containerd/containerd/v2/core/remotes/docker"
	"github.com/containerd/containerd/v2/pkg/cap"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/oci"
	"github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/containerutil"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/containerd/nerdctl/v2/pkg/referenceutil"
	"github.com/containerd/platforms"
	"github.com/distribution/reference"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

//go:generate mockgen --destination=../../mocks/mocks_backend/containerdclient.go -package=mocks_backend github.com/runfinch/finch-daemon/internal/backend ContainerdClient
type ContainerdClient interface {
	GetClient() *containerd.Client
	GetContainerStatus(ctx context.Context, c containerd.Container) containerd.ProcessStatus
	SearchContainer(ctx context.Context, searchText string) (containers []containerd.Container, err error)
	GetImage(ctx context.Context, ref string) (containerd.Image, error)
	SearchImage(ctx context.Context, searchText string) ([]images.Image, error)
	ParsePlatform(platform string) (ocispec.Platform, error)
	DefaultPlatformSpec() ocispec.Platform
	DefaultPlatformStrict() platforms.MatchComparer
	ParseDockerRef(rawRef string) (ref, refDomain string, err error)
	DefaultDockerHost(refDomain string) (string, error)
	GetContainerTaskWait(ctx context.Context, attach cio.Attach, c containerd.Container) (task containerd.Task, waitCh <-chan containerd.ExitStatus, err error)
	GetContainerRemoveEvent(ctx context.Context, c containerd.Container) (<-chan *events.Envelope, <-chan error)
	ListSnapshotMounts(ctx context.Context, cid string) ([]mount.Mount, error)
	MountAll(mounts []mount.Mount, mPath string) error
	Unmount(mPath string, flags int) error
	ImageService() images.Store
	ConvertImage(ctx context.Context, dstRef, srcRef string, opts ...converter.Opt) (*images.Image, error)
	DeleteImage(ctx context.Context, img string) error
	GetImageDigests(ctx context.Context, img *images.Image) (digests []digest.Digest, err error)
	GetUsedImages(ctx context.Context) (stopped, running map[string]string, err error)
	OCISpecWithUser(user string) oci.SpecOpts
	OCISpecWithAdditionalGIDs(user string) oci.SpecOpts
	GetCurrentCapabilities() ([]string, error)
	NewFIFOSetInDir(root, id string, terminal bool) (*cio.FIFOSet, error)
	NewDirectCIO(ctx context.Context, fifos *cio.FIFOSet) (*cio.DirectIO, error)
	SubscribeToEvents(ctx context.Context, filters ...string) (<-chan *events.Envelope, <-chan error)
	PublishEvent(ctx context.Context, topic string, event events.Event) error
}

type ContainerdClientWrapper struct {
	client *containerd.Client
}

// NewContainerdClientWrapper creates a new instance of ContainerdClientWrapper.
func NewContainerdClientWrapper(client *containerd.Client) *ContainerdClientWrapper {
	return &ContainerdClientWrapper{
		client: client,
	}
}

func (w *ContainerdClientWrapper) GetClient() *containerd.Client {
	return w.client
}

// GetContainerStatus wraps the containerd function to get the status of a container.
func (w *ContainerdClientWrapper) GetContainerStatus(ctx context.Context, c containerd.Container) containerd.ProcessStatus {
	// Just in case, there is something wrong in server.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	task, err := c.Task(ctx, nil)
	if err != nil {
		// NOTE: NotFound doesn't mean that container hasn't started.
		// In docker/CRI-containerd plugin, the task will be deleted
		// when it exits. So, the status will be "created" for this
		// case.
		if errdefs.IsNotFound(err) {
			return containerd.Created
		}
		return containerd.Unknown
	}

	status, err := task.Status(ctx)
	if err != nil {
		return containerd.Unknown
	}
	return status.Status
}

// SearchContainer returns the list of containers that match the prefix.
func (w *ContainerdClientWrapper) SearchContainer(ctx context.Context, searchText string) (containers []containerd.Container, err error) {
	filters := []string{
		fmt.Sprintf("labels.%q==%s", labels.Name, searchText),
		fmt.Sprintf("id~=^%s.*$", regexp.QuoteMeta(searchText)),
	}

	containers, err = w.client.Containers(ctx, filters...)
	return containers, err
}

// GetImage returns an image with given reference.
func (w *ContainerdClientWrapper) GetImage(ctx context.Context, ref string) (containerd.Image, error) {
	return w.client.GetImage(ctx, ref)
}

// SearchImage returns a list of images that match the search prefix.
func (w *ContainerdClientWrapper) SearchImage(ctx context.Context, searchText string) ([]images.Image, error) {
	var filters []string
	if canonicalRef, err := referenceutil.Parse(searchText); err == nil {
		filters = append(filters, fmt.Sprintf("name==%s", canonicalRef.String()))
	}
	filters = append(filters,
		fmt.Sprintf("name==%s", searchText),
		fmt.Sprintf("target.digest~=^sha256:%s.*$", regexp.QuoteMeta(searchText)),
		fmt.Sprintf("target.digest~=^%s.*$", regexp.QuoteMeta(searchText)),
	)

	return w.client.ImageService().List(ctx, filters...)
}

// ParsePlatform parses a platform text into an ocispec Platform type.
func (*ContainerdClientWrapper) ParsePlatform(platform string) (ocispec.Platform, error) {
	return platforms.Parse(platform)
}

// DefaultPlatformSpec returns the current platform's default platform specification.
func (w *ContainerdClientWrapper) DefaultPlatformSpec() ocispec.Platform {
	return platforms.DefaultSpec()
}

// DefaultPlatformStrict returns the strict form of current platform's default platform specification.
func (w *ContainerdClientWrapper) DefaultPlatformStrict() platforms.MatchComparer {
	return platforms.DefaultStrict()
}

// ParseDockerRef normalizes the image reference following the docker convention.
func (w *ContainerdClientWrapper) ParseDockerRef(rawRef string) (ref, refDomain string, err error) {
	named, err := reference.ParseDockerRef(rawRef)
	if err != nil {
		return
	}
	ref = named.String()
	refDomain = reference.Domain(named)
	return
}

// DefaultDockerHost converts "docker.io" to "registry-1.docker.io".
func (w *ContainerdClientWrapper) DefaultDockerHost(refDomain string) (string, error) {
	return docker.DefaultHost(refDomain)
}

// GetContainerTaskWait gets the wait channel for a container in the process of doing a task.
func (*ContainerdClientWrapper) GetContainerTaskWait(ctx context.Context, attach cio.Attach, c containerd.Container) (task containerd.Task, waitCh <-chan containerd.ExitStatus, err error) {
	task, err = c.Task(ctx, attach)
	if err != nil {
		waitCh = nil
		return
	}
	waitCh, err = task.Wait(ctx)
	return
}

// GetContainerRemoveEvent subscribes to the remove event for the given container and returns its channel.
func (w *ContainerdClientWrapper) GetContainerRemoveEvent(ctx context.Context, c containerd.Container) (<-chan *events.Envelope, <-chan error) {
	return w.client.Subscribe(ctx,
		fmt.Sprintf(`topic=="/containers/delete",event.id=="%s"`, c.ID()),
	)
}

func (w *ContainerdClientWrapper) ListSnapshotMounts(ctx context.Context, key string) ([]mount.Mount, error) {
	return w.client.SnapshotService("").Mounts(ctx, key)
}

func (*ContainerdClientWrapper) MountAll(mounts []mount.Mount, mPath string) error {
	return mount.All(mounts, mPath)
}

func (*ContainerdClientWrapper) Unmount(mPath string, flags int) error {
	return mount.Unmount(mPath, flags)
}

func (w *ContainerdClientWrapper) ImageService() images.Store {
	return w.client.ImageService()
}

func (w *ContainerdClientWrapper) ConvertImage(ctx context.Context, dstRef, srcRef string, opts ...converter.Opt) (*images.Image, error) {
	return converter.Convert(ctx, w.client, dstRef, srcRef, opts...)
}

// DeleteImage deletes an image.
func (w *ContainerdClientWrapper) DeleteImage(ctx context.Context, img string) error {
	return w.client.ImageService().Delete(ctx, img, images.SynchronousDelete())
}

// GetImageDigests returns the list of digests for a given image.
func (w *ContainerdClientWrapper) GetImageDigests(ctx context.Context, img *images.Image) (digests []digest.Digest, err error) {
	cntStore := w.client.ContentStore()
	return img.RootFS(ctx, cntStore, platforms.DefaultStrict())
}

// GetUsedImages returns the list of images that are used by containers.
// `stopped` contains the images used by stopped containers, `running` contains the images used by running containers.
func (w *ContainerdClientWrapper) GetUsedImages(ctx context.Context) (stopped, running map[string]string, err error) {
	stopped = make(map[string]string)
	running = make(map[string]string)
	containerList, err := w.client.Containers(ctx)
	if err != nil {
		return
	}
	for _, cont := range containerList {
		image, err := cont.Image(ctx)
		// skip if the image is not found
		if err != nil {
			continue
		}
		switch cStatus, _ := containerutil.ContainerStatus(ctx, cont); cStatus.Status {
		case containerd.Running, containerd.Pausing, containerd.Paused:
			running[image.Name()] = cont.ID()
		default:
			stopped[image.Name()] = cont.ID()
		}
	}
	return stopped, running, err
}

func (*ContainerdClientWrapper) OCISpecWithUser(user string) oci.SpecOpts {
	return oci.WithUser(user)
}

func (*ContainerdClientWrapper) OCISpecWithAdditionalGIDs(user string) oci.SpecOpts {
	return oci.WithAdditionalGIDs(user)
}

func (*ContainerdClientWrapper) GetCurrentCapabilities() ([]string, error) {
	return cap.Current()
}

func (*ContainerdClientWrapper) NewFIFOSetInDir(root, id string, terminal bool) (*cio.FIFOSet, error) {
	return cio.NewFIFOSetInDir(root, id, terminal)
}

func (*ContainerdClientWrapper) NewDirectCIO(ctx context.Context, fifos *cio.FIFOSet) (*cio.DirectIO, error) {
	return cio.NewDirectIO(ctx, fifos)
}

func (w *ContainerdClientWrapper) SubscribeToEvents(ctx context.Context, filters ...string) (<-chan *events.Envelope, <-chan error) {
	return w.client.EventService().Subscribe(ctx, filters...)
}

func (w *ContainerdClientWrapper) PublishEvent(ctx context.Context, topic string, event events.Event) error {
	return w.client.EventService().Publish(ctx, topic, event)
}
