// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	"github.com/runfinch/finch-daemon/api/types"
)

func (s *service) Inspect(ctx context.Context, cid string, sizeFlag bool) (*types.Container, error) {
	c, err := s.getContainer(ctx, cid)
	if err != nil {
		return nil, err
	}

	inspect, err := s.nctlContainerSvc.InspectContainer(ctx, c, sizeFlag)
	if err != nil {
		return nil, err
	}

	// translate to a finch-daemon container inspect type
	cont := types.Container{
		ID:              inspect.ID,
		Created:         inspect.Created,
		Path:            inspect.Path,
		Args:            inspect.Args,
		State:           inspect.State,
		Image:           inspect.Image,
		ResolvConfPath:  inspect.ResolvConfPath,
		HostnamePath:    inspect.HostnamePath,
		LogPath:         inspect.LogPath,
		Name:            fmt.Sprintf("/%s", inspect.Name),
		RestartCount:    inspect.RestartCount,
		Driver:          inspect.Driver,
		Platform:        inspect.Platform,
		AppArmorProfile: inspect.AppArmorProfile,
		Mounts:          inspect.Mounts,
		NetworkSettings: inspect.NetworkSettings,
		SizeRw:          inspect.SizeRw,
		SizeRootFs:      inspect.SizeRootFs,
	}

	cont.Config = &types.ContainerConfig{
		Hostname:     inspect.Config.Hostname,
		User:         inspect.Config.User,
		AttachStdin:  inspect.Config.AttachStdin,
		ExposedPorts: inspect.Config.ExposedPorts,
		Tty:          false, // TODO: Tty is always false until attach supports stdin with tty
		Env:          inspect.Config.Env,
		Cmd:          inspect.Config.Cmd,
		Image:        inspect.Image,
		Volumes:      inspect.Config.Volumes,
		WorkingDir:   inspect.Config.WorkingDir,
		Entrypoint:   inspect.Config.Entrypoint,
		Labels:       inspect.Config.Labels,
	}

	// make sure it passes the default time value for time fields otherwise the goclient fails.
	if inspect.Created == "" {
		cont.Created = "0001-01-01T00:00:00Z"
	}

	if inspect.State != nil && inspect.State.FinishedAt == "" {
		cont.State.FinishedAt = "0001-01-01T00:00:00Z"
	}

	return &cont, nil
}
