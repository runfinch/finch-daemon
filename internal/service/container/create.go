// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"encoding/json"
	"fmt"

	containerd "github.com/containerd/containerd/v2/client"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/sirupsen/logrus"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Create(ctx context.Context, image string, cmd []string, createOpt types.ContainerCreateOptions, netOpt types.NetworkOptions) (cid string, err error) {
	// Set path to nerdctl binary required for OCI hooks and logging
	if createOpt.NerdctlCmd == "" {
		ncExe, err := s.nctlContainerSvc.GetNerdctlExe()
		if err != nil {
			return "", fmt.Errorf("failed to find nerdctl binary: %s", err)
		}
		createOpt.NerdctlCmd = ncExe
		createOpt.NerdctlArgs = []string{}
	}

	netManager, err := s.nctlContainerSvc.NewNetworkingOptionsManager(netOpt)
	if err != nil {
		logrus.Debugf("error creating network manager for the given network options: %s", err)
		return "", err
	}

	args := []string{image}
	args = append(args, cmd...)
	cont, gc, err := s.nctlContainerSvc.CreateContainer(ctx, args, netManager, createOpt)
	if err != nil {
		if gc != nil {
			gc()
		}
		logrus.Debugf("failed to create container: %s", err)

		// translate error definitions from containerd
		switch {
		case cerrdefs.IsNotFound(err):
			return "", errdefs.NewNotFound(err)
		case cerrdefs.IsInvalidArgument(err):
			return "", errdefs.NewInvalidFormat(err)
		case cerrdefs.IsAlreadyExists(err):
			return "", errdefs.NewConflict(err)
		default:
			return "", err
		}
	}

	updateContainerMetadata(ctx, createOpt, netOpt, cont)

	return cont.ID(), nil
}

func updateContainerMetadata(ctx context.Context, createOpt types.ContainerCreateOptions, netOpt types.NetworkOptions, cont containerd.Container) error {
	// get container labels
	opts, err := cont.Labels(ctx)
	if err != nil {
		logrus.Errorf("failed to get container labels: %s", err)
		return err
	}
	// get oci spec
	spec, err := cont.Spec(ctx)
	if err != nil {
		logrus.Errorf("failed to get container OCI spec: %s", err)
		return err
	}

	// Note: OCI hooks are NOT stripped here. Containers created via the HTTP API
	// may be started either via the HTTP API (customStart strips hooks and runs
	// setupNetworking inline) or via nerdctl CLI (which needs hooks intact for
	// runc to execute). Stripping hooks is done in customStart instead.

	// Handle port labels for backward compatibility with nerdctl 2.1.2.
	// nerdctl 2.1.3 changed the port publishing logic to use a dedicated portstore instead of container labels.
	// Though nerdctl itself is backward compatible, newer library versions will no longer create the port labels,
	// which is not compatible with older nerdctl executables needed to run the OCI hooks for setting up the network.
	if len(netOpt.PortMappings) > 0 {
		portsJSON, err := json.Marshal(netOpt.PortMappings)
		if err != nil {
			return err
		}
		opts[labels.Ports] = string(portsJSON)
		spec.Annotations[labels.Ports] = string(portsJSON)
	}

	err = cont.Update(ctx,
		containerd.UpdateContainerOpts(containerd.WithContainerLabels(opts)),
		containerd.UpdateContainerOpts(containerd.WithSpec(spec)),
	)
	if err != nil {
		logrus.Errorf("failed to update container: %s", err)
		return err
	}

	return nil
}
