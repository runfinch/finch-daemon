// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/clientutil"
	"github.com/containerd/nerdctl/v2/pkg/cmd/container"
	"github.com/containerd/nerdctl/v2/pkg/containerinspector"
	"github.com/containerd/nerdctl/v2/pkg/containerutil"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	"github.com/containerd/nerdctl/v2/pkg/logging"
)

//go:generate mockgen --destination=../../mocks/mocks_backend/nerdctlcontainersvc.go -package=mocks_backend github.com/runfinch/finch-daemon/internal/backend NerdctlContainerSvc
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

func (w *NerdctlWrapper) RemoveContainer(ctx context.Context, c containerd.Container, force bool, removeVolumes bool) error {
	return container.RemoveContainer(ctx, c, *w.globalOptions, force, removeVolumes, w.clientWrapper.client)
}

// StartContainer wrapper function to call nerdctl function to start a container.
func (w *NerdctlWrapper) StartContainer(ctx context.Context, container containerd.Container) error {
	return containerutil.Start(ctx, container, false, w.clientWrapper.client, "")
}

// StopContainer wrapper function to call nerdctl function to stop a container.
func (*NerdctlWrapper) StopContainer(ctx context.Context, container containerd.Container, timeout *time.Duration) error {
	return containerutil.Stop(ctx, container, timeout)
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

func (w *NerdctlWrapper) ListContainers(ctx context.Context, options types.ContainerListOptions) ([]container.ListItem, error) {
	return container.List(ctx, w.clientWrapper.client, options)
}

func (w *NerdctlWrapper) RenameContainer(ctx context.Context, con containerd.Container, newName string, options types.ContainerRenameOptions) error {
	return container.Rename(ctx, w.clientWrapper.client, con.ID(), newName, options)
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
