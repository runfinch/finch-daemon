// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package backend uses interfaces and structs to create abstractions for nerdctl and containerd function calls,
// which allows mock creation for unit testing using mockgen.
package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/images/converter"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/pkg/cap"
	"github.com/containerd/containerd/platforms"
	refdocker "github.com/containerd/containerd/reference/docker"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	dockerconfig "github.com/containerd/containerd/remotes/docker/config"
	"github.com/containerd/nerdctl/pkg/api/types"
	"github.com/containerd/nerdctl/pkg/buildkitutil"
	"github.com/containerd/nerdctl/pkg/clientutil"
	"github.com/containerd/nerdctl/pkg/cmd/builder"
	"github.com/containerd/nerdctl/pkg/cmd/container"
	"github.com/containerd/nerdctl/pkg/cmd/volume"
	"github.com/containerd/nerdctl/pkg/containerinspector"
	"github.com/containerd/nerdctl/pkg/containerutil"
	"github.com/containerd/nerdctl/pkg/idutil/imagewalker"
	"github.com/containerd/nerdctl/pkg/imageinspector"
	"github.com/containerd/nerdctl/pkg/imgutil"
	"github.com/containerd/nerdctl/pkg/imgutil/dockerconfigresolver"
	"github.com/containerd/nerdctl/pkg/imgutil/push"
	"github.com/containerd/nerdctl/pkg/infoutil"
	"github.com/containerd/nerdctl/pkg/inspecttypes/dockercompat"
	"github.com/containerd/nerdctl/pkg/inspecttypes/native"
	"github.com/containerd/nerdctl/pkg/labels"
	"github.com/containerd/nerdctl/pkg/logging"
	"github.com/containerd/nerdctl/pkg/netutil"
	"github.com/containerd/nerdctl/pkg/referenceutil"
	"github.com/containernetworking/cni/libcni"
	"github.com/containernetworking/cni/pkg/invoke"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

//go:generate mockgen --destination=../mocks/mocks_backend/nerdctlimagesvc.go -package=mocks_backend github.com/runfinch/finch-daemon/pkg/backend NerdctlImageSvc
type NerdctlImageSvc interface {
	InspectImage(ctx context.Context, image images.Image) (*dockercompat.Image, error)
	GetDockerResolver(ctx context.Context, refDomain string, creds dockerconfigresolver.AuthCreds) (remotes.Resolver, docker.StatusTracker, error)
	PullImage(ctx context.Context, stdout, stderr io.Writer, resolver remotes.Resolver, ref string, platforms []ocispec.Platform) (*imgutil.EnsuredImage, error)
	PushImage(ctx context.Context, resolver remotes.Resolver, tracker docker.StatusTracker, stdout io.Writer, pushRef, ref string, platMC platforms.MatchComparer) error
	SearchImage(ctx context.Context, name string) (int, int, []*images.Image, error)
}

//go:generate mockgen --destination=../mocks/mocks_backend/nerdctlbuildersvc.go -package=mocks_backend github.com/runfinch/finch-daemon/pkg/backend NerdctlBuilderSvc
type NerdctlBuilderSvc interface {
	Build(ctx context.Context, client ContainerdClient, options types.BuilderBuildOptions) error
	GetBuildkitHost() (string, error)
}

//go:generate mockgen --destination=../mocks/mocks_backend/nerdctlcontainersvc.go -package=mocks_backend github.com/runfinch/finch-daemon/pkg/backend NerdctlContainerSvc
type NerdctlContainerSvc interface {
	RemoveContainer(ctx context.Context, c containerd.Container, force bool, removeAnonVolumes bool) error
	StartContainer(ctx context.Context, container containerd.Container) error
	StopContainer(ctx context.Context, container containerd.Container, timeout *time.Duration) error
	CreateContainer(ctx context.Context, args []string, netManager containerutil.NetworkOptionsManager, options types.ContainerCreateOptions) (containerd.Container, func(), error)
	InspectContainer(ctx context.Context, c containerd.Container) (*dockercompat.Container, error)
	InspectNetNS(ctx context.Context, pid int) (*native.NetNS, error)
	NewNetworkingOptionsManager(types.NetworkOptions) (containerutil.NetworkOptionsManager, error)
	ListContainers(ctx context.Context, options types.ContainerListOptions) ([]container.ListItem, error)
	RenameContainer(ctx context.Context, container containerd.Container, newName string, options types.ContainerRenameOptions) error

	// Mocked functions for container attach
	GetDataStore() (string, error)
	LoggingInitContainerLogViewer(containerLabels map[string]string, lvopts logging.LogViewOptions, stopChannel chan os.Signal, experimental bool) (contlv *logging.ContainerLogViewer, err error)
	LoggingPrintLogsTo(stdout, stderr io.Writer, clv *logging.ContainerLogViewer) error

	// GetNerdctlExe returns a path to the nerdctl binary, which is required for setting up OCI hooks and logging
	GetNerdctlExe() (string, error)
}

//go:generate mockgen --destination=../mocks/mocks_backend/nerdctlnetworksvc.go -package=mocks_backend github.com/runfinch/finch-daemon/pkg/backend NerdctlNetworkSvc
type NerdctlNetworkSvc interface {
	FilterNetworks(filterf func(networkConfig *netutil.NetworkConfig) bool) ([]*netutil.NetworkConfig, error)
	AddNetworkList(ctx context.Context, netconflist *libcni.NetworkConfigList, conf *libcni.RuntimeConf) (cnitypes.Result, error)
	CreateNetwork(opts netutil.CreateOptions) (*netutil.NetworkConfig, error)
	RemoveNetwork(networkConfig *netutil.NetworkConfig) error
	InspectNetwork(ctx context.Context, networkConfig *netutil.NetworkConfig) (*dockercompat.Network, error)
	UsedNetworkInfo(ctx context.Context) (map[string][]string, error)
	NetconfPath() string
}

//go:generate mockgen --destination=../mocks/mocks_backend/nerdctlvolumesvc.go -package=mocks_backend github.com/runfinch/finch-daemon/pkg/backend NerdctlVolumeSvc
type NerdctlVolumeSvc interface {
	ListVolumes(size bool, filters []string) (map[string]native.Volume, error)
	RemoveVolume(ctx context.Context, name string, force bool, stdout io.Writer) error
	GetVolume(name string) (*native.Volume, error)
	CreateVolume(name string, labels []string) (*native.Volume, error)
}

//go:generate mockgen --destination=../mocks/mocks_backend/nerdctlsystemsvc.go -package=mocks_backend github.com/runfinch/finch-daemon/pkg/backend NerdctlSystemSvc
type NerdctlSystemSvc interface {
	GetServerVersion(ctx context.Context) (*dockercompat.ServerVersion, error)
}

type NerdctlWrapper struct {
	clientWrapper *ContainerdClientWrapper
	globalOptions *types.GlobalCommandOptions
	nerdctlExe    string
	netClient     *netutil.CNIEnv
	CNI           *libcni.CNIConfig
}

func NewNerdctlWrapper(clientWrapper *ContainerdClientWrapper, options *types.GlobalCommandOptions) *NerdctlWrapper {
	return &NerdctlWrapper{
		clientWrapper: clientWrapper,
		globalOptions: options,
		netClient: &netutil.CNIEnv{
			Path:        options.CNIPath,
			NetconfPath: options.CNINetConfPath,
		},
		CNI: libcni.NewCNIConfig(
			[]string{
				options.CNIPath,
			},
			&invoke.DefaultExec{
				RawExec:       &invoke.RawExec{Stderr: os.Stderr},
				PluginDecoder: version.PluginDecoder{},
			}),
	}
}

func (w *NerdctlWrapper) GetNerdctlExe() (string, error) {
	if w.nerdctlExe != "" {
		return w.nerdctlExe, nil
	}
	exe, err := exec.LookPath("nerdctl")
	if err != nil {
		return "", err
	}
	w.nerdctlExe = exe
	return exe, nil
}

func (w *NerdctlWrapper) CreateVolume(name string, labels []string) (*native.Volume, error) {
	volumeCreateOpts := types.VolumeCreateOptions{
		Stdout:   os.Stdout,
		Labels:   labels,
		GOptions: *w.globalOptions,
	}
	return volume.Create(name, volumeCreateOpts)
}

func (w *NerdctlWrapper) ListVolumes(size bool, filters []string) (map[string]native.Volume, error) {
	vols, err := volume.Volumes(
		w.globalOptions.Namespace,
		w.globalOptions.DataRoot,
		w.globalOptions.Address,
		size,
		filters,
	)
	if err != nil {
		return nil, err
	}
	return vols, err
}

// GetVolume wrapper function to call nerdctl function to get the details of a volume
func (w *NerdctlWrapper) GetVolume(name string) (*native.Volume, error) {
	volStore, err := volume.Store(w.globalOptions.Namespace, w.globalOptions.DataRoot, w.globalOptions.Address)
	if err != nil {
		return nil, err
	}
	vols, err := volStore.Get(name, false)
	if err != nil {
		return nil, err
	}
	return vols, err
}

// RemoveVolume wrapper function to call nerdctl function to remove a volume
func (w *NerdctlWrapper) RemoveVolume(ctx context.Context, name string, force bool, stdout io.Writer) error {
	return volume.Remove(
		ctx,
		w.clientWrapper.client,
		[]string{name},
		types.VolumeRemoveOptions{
			Stdout:   stdout,
			GOptions: *w.globalOptions,
			Force:    force,
		})
}

func (w *NerdctlWrapper) InspectImage(ctx context.Context, image images.Image) (*dockercompat.Image, error) {
	n, err := imageinspector.Inspect(ctx, w.clientWrapper.client, image, w.globalOptions.Snapshotter)
	if err != nil {
		return nil, err
	}
	return dockercompat.ImageFromNative(n)
}

// GetDockerResolver returns a new Docker config resolver from the reference host and auth credentials
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

// PullImage pulls an image from nerdctl's imgutil library
func (w *NerdctlWrapper) PullImage(ctx context.Context, stdout, stderr io.Writer, resolver remotes.Resolver, ref string, platforms []ocispec.Platform) (*imgutil.EnsuredImage, error) {
	return imgutil.PullImage(
		ctx,
		w.clientWrapper.client,
		stdout, stderr,
		w.globalOptions.Snapshotter,
		resolver,
		ref,
		platforms,
		nil,
		false,
		imgutil.RemoteSnapshotterFlags{},
	)
}

// PushImage pushes an image using nerdctl's imgutil library
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

func (w *NerdctlWrapper) RemoveContainer(ctx context.Context, c containerd.Container, force bool, removeVolumes bool) error {
	return container.RemoveContainer(ctx, c, *w.globalOptions, force, removeVolumes, w.clientWrapper.client)
}

// StartContainer wrapper function to call nerdctl function to start a container
func (w *NerdctlWrapper) StartContainer(ctx context.Context, container containerd.Container) error {
	return containerutil.Start(ctx, container, false, w.clientWrapper.client, "")
}

// StopContainer wrapper function to call nerdctl function to stop a container
func (*NerdctlWrapper) StopContainer(ctx context.Context, container containerd.Container, timeout *time.Duration) error {
	return containerutil.Stop(ctx, container, timeout)
}

func (*NerdctlWrapper) Build(ctx context.Context, client ContainerdClient, options types.BuilderBuildOptions) error {
	return builder.Build(ctx, client.GetClient(), options)
}

func (w *NerdctlWrapper) GetBuildkitHost() (string, error) {
	return buildkitutil.GetBuildkitHost(w.globalOptions.Namespace)
}

func (w *NerdctlWrapper) CreateContainer(ctx context.Context, args []string, netManager containerutil.NetworkOptionsManager, options types.ContainerCreateOptions) (containerd.Container, func(), error) {
	return container.Create(ctx, w.clientWrapper.client, args, netManager, options)
}

func (w *NerdctlWrapper) InspectContainer(ctx context.Context, c containerd.Container) (*dockercompat.Container, error) {
	n, err := containerinspector.Inspect(ctx, c)
	if err != nil {
		return nil, err
	}
	return dockercompat.ContainerFromNative(n)
}

func (w *NerdctlWrapper) InspectNetNS(ctx context.Context, pid int) (*native.NetNS, error) {
	return containerinspector.InspectNetNS(ctx, pid)
}

func (w *NerdctlWrapper) NewNetworkingOptionsManager(options types.NetworkOptions) (containerutil.NetworkOptionsManager, error) {
	return containerutil.NewNetworkingOptionsManager(*w.globalOptions, options, w.clientWrapper.client)
}

func (w *NerdctlWrapper) FilterNetworks(filterf func(networkConfig *netutil.NetworkConfig) bool) ([]*netutil.NetworkConfig, error) {
	return w.netClient.FilterNetworks(filterf)
}

func (w *NerdctlWrapper) AddNetworkList(ctx context.Context, netconflist *libcni.NetworkConfigList, conf *libcni.RuntimeConf) (cnitypes.Result, error) {
	return w.CNI.AddNetworkList(ctx, netconflist, conf)
}

func (w *NerdctlWrapper) CreateNetwork(opts netutil.CreateOptions) (*netutil.NetworkConfig, error) {
	return w.netClient.CreateNetwork(opts)
}

func (w *NerdctlWrapper) RemoveNetwork(networkConfig *netutil.NetworkConfig) error {
	return w.netClient.RemoveNetwork(networkConfig)
}

func (w *NerdctlWrapper) InspectNetwork(ctx context.Context, networkConfig *netutil.NetworkConfig) (*dockercompat.Network, error) {
	network := &native.Network{
		CNI:           json.RawMessage(networkConfig.Bytes),
		NerdctlID:     networkConfig.NerdctlID,
		NerdctlLabels: networkConfig.NerdctlLabels,
		File:          networkConfig.File,
	}
	return dockercompat.NetworkFromNative(network)
}

func (w *NerdctlWrapper) UsedNetworkInfo(ctx context.Context) (map[string][]string, error) {
	return netutil.UsedNetworks(ctx, w.clientWrapper.client)
}

func (w *NerdctlWrapper) NetconfPath() string {
	return w.netClient.NetconfPath
}

func (w *NerdctlWrapper) GetDataStore() (string, error) {
	return clientutil.DataStore(w.globalOptions.DataRoot, w.globalOptions.Address)
}

func (*NerdctlWrapper) LoggingInitContainerLogViewer(containerLabels map[string]string, lvopts logging.LogViewOptions, stopChannel chan os.Signal, experimental bool) (contlv *logging.ContainerLogViewer, err error) {
	return logging.InitContainerLogViewer(containerLabels, lvopts, stopChannel, experimental)
}

func (*NerdctlWrapper) LoggingPrintLogsTo(stdout, stderr io.Writer, clv *logging.ContainerLogViewer) error {
	return clv.PrintLogsTo(stdout, stderr)
}

func (w *NerdctlWrapper) ListContainers(ctx context.Context, options types.ContainerListOptions) ([]container.ListItem, error) {
	return container.List(ctx, w.clientWrapper.client, options)
}

func (w *NerdctlWrapper) RenameContainer(ctx context.Context, con containerd.Container, newName string, options types.ContainerRenameOptions) error {
	return container.Rename(ctx, w.clientWrapper.client, con.ID(), newName, options)
}

func (w *NerdctlWrapper) GetServerVersion(ctx context.Context) (*dockercompat.ServerVersion, error) {
	return infoutil.ServerVersion(ctx, w.clientWrapper.GetClient())
}

//go:generate mockgen --destination=../mocks/mocks_backend/containerdclient.go -package=mocks_backend github.com/runfinch/finch-daemon/pkg/backend ContainerdClient
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

// NewContainerdClientWrapper creates a new instance of ContainerdClientWrapper
func NewContainerdClientWrapper(client *containerd.Client) *ContainerdClientWrapper {
	return &ContainerdClientWrapper{
		client: client,
	}
}

func (w *ContainerdClientWrapper) GetClient() *containerd.Client {
	return w.client
}

// GetContainerStatus wraps the containerd function to get the status of a container
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

// SearchContainer returns the list of containers that match the prefix
func (w *ContainerdClientWrapper) SearchContainer(ctx context.Context, searchText string) (containers []containerd.Container, err error) {
	filters := []string{
		fmt.Sprintf("labels.%q==%s", labels.Name, searchText),
		fmt.Sprintf("id~=^%s.*$", regexp.QuoteMeta(searchText)),
	}

	containers, err = w.client.Containers(ctx, filters...)
	return containers, err
}

// GetImage returns an image with given reference
func (w *ContainerdClientWrapper) GetImage(ctx context.Context, ref string) (containerd.Image, error) {
	return w.client.GetImage(ctx, ref)
}

// SearchImage returns a list of images that match the search prefix
func (w *ContainerdClientWrapper) SearchImage(ctx context.Context, searchText string) ([]images.Image, error) {
	var filters []string
	if canonicalRef, err := referenceutil.ParseAny(searchText); err == nil {
		filters = append(filters, fmt.Sprintf("name==%s", canonicalRef.String()))
	}
	filters = append(filters,
		fmt.Sprintf("name==%s", searchText),
		fmt.Sprintf("target.digest~=^sha256:%s.*$", regexp.QuoteMeta(searchText)),
		fmt.Sprintf("target.digest~=^%s.*$", regexp.QuoteMeta(searchText)),
	)

	return w.client.ImageService().List(ctx, filters...)
}

// ParsePlatform parses a platform text into an ocispec Platform type
func (*ContainerdClientWrapper) ParsePlatform(platform string) (ocispec.Platform, error) {
	return platforms.Parse(platform)
}

// DefaultPlatformSpec returns the current platform's default platform specification
func (w *ContainerdClientWrapper) DefaultPlatformSpec() ocispec.Platform {
	return platforms.DefaultSpec()
}

// DefaultPlatformStrict returns the strict form of current platform's default platform specification
func (w *ContainerdClientWrapper) DefaultPlatformStrict() platforms.MatchComparer {
	return platforms.DefaultStrict()
}

// ParseDockerRef normalizes the image reference following the docker convention
func (w *ContainerdClientWrapper) ParseDockerRef(rawRef string) (ref, refDomain string, err error) {
	named, err := refdocker.ParseDockerRef(rawRef)
	if err != nil {
		return
	}
	ref = named.String()
	refDomain = refdocker.Domain(named)
	return
}

// DefaultDockerHost converts "docker.io" to "registry-1.docker.io"
func (w *ContainerdClientWrapper) DefaultDockerHost(refDomain string) (string, error) {
	return docker.DefaultHost(refDomain)
}

// GetContainerTaskWait gets the wait channel for a container in the process of doing a task
func (*ContainerdClientWrapper) GetContainerTaskWait(ctx context.Context, attach cio.Attach, c containerd.Container) (task containerd.Task, waitCh <-chan containerd.ExitStatus, err error) {
	task, err = c.Task(ctx, attach)
	if err != nil {
		waitCh = nil
		return
	}
	waitCh, err = task.Wait(ctx)
	return
}

// GetContainerRemoveEvent subscribes to the remove event for the given container and returns its channel
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

// DeleteImage deletes an image
func (w *ContainerdClientWrapper) DeleteImage(ctx context.Context, img string) error {
	return w.client.ImageService().Delete(ctx, img, images.SynchronousDelete())
}

// GetImageDigests returns the list of digests for a given image
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

func (c *ContainerdClientWrapper) SubscribeToEvents(ctx context.Context, filters ...string) (<-chan *events.Envelope, <-chan error) {
	return c.client.EventService().Subscribe(ctx, filters...)
}

func (c *ContainerdClientWrapper) PublishEvent(ctx context.Context, topic string, event events.Event) error {
	return c.client.EventService().Publish(ctx, topic, event)
}
