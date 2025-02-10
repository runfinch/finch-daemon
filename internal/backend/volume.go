// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"
	"io"
	"os"

	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/cmd/volume"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
)

//go:generate mockgen --destination=../../mocks/mocks_backend/nerdctlvolumesvc.go -package=mocks_backend github.com/runfinch/finch-daemon/internal/backend NerdctlVolumeSvc
type NerdctlVolumeSvc interface {
	ListVolumes(size bool, filters []string) (map[string]native.Volume, error)
	RemoveVolume(ctx context.Context, name string, force bool, stdout io.Writer) error
	GetVolume(name string) (*native.Volume, error)
	CreateVolume(name string, labels []string) (*native.Volume, error)
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

// GetVolume wrapper function to call nerdctl function to get the details of a volume.
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

// RemoveVolume wrapper function to call nerdctl function to remove a volume.
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
