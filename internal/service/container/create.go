// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/clientutil"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/containerd/nerdctl/v2/pkg/logging"
	"github.com/containerd/nerdctl/v2/pkg/netutil"
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

	// translate network IDs to names because nerdctl currently does not recognize networks by their IDs during create.
	// TODO: remove this when the issue is fixed upstream.
	if err := s.translateNetworkIds(&netOpt); err != nil {
		return "", err
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

	// NOTE: this is a temporary workaround to fix logging issue described in https://github.com/containerd/nerdctl/issues/2264.
	// The refactored create method in nerdctl uses self exe (finch-daemon) binary for logging instead of nerdctl binary path.
	// The following workaround resets this logging binary in the OCI spec.
	// TODO: remove this workaround when the issue is resolved upstream.
	resetLogURI(ctx, createOpt, cont)

	return cont.ID(), nil
}

// translateNetworkIds translates network IDs to corresponding network names in network options.
func (s *service) translateNetworkIds(netOpt *types.NetworkOptions) error {
	for i, netId := range netOpt.NetworkSlice {
		if netId == "host" || netId == "none" || netId == "bridge" {
			continue
		}

		netList, err := s.nctlContainerSvc.FilterNetworks(func(networkConfig *netutil.NetworkConfig) bool {
			return networkConfig.Name == netId || *networkConfig.NerdctlID == netId
		})
		if err != nil {
			return err
		}
		if len(netList) == 0 {
			return errdefs.NewNotFound(fmt.Errorf("network not found: %s", netId))
		} else if len(netList) > 1 {
			return fmt.Errorf("multiple networks found for id: %s", netId)
		}
		netOpt.NetworkSlice[i] = netList[0].Name
	}

	return nil
}

func resetLogURI(ctx context.Context, createOpt types.ContainerCreateOptions, cont containerd.Container) error {
	// get data store directory for logging
	dataStore, err := clientutil.DataStore(createOpt.GOptions.DataRoot, createOpt.GOptions.Address)
	if err != nil {
		logrus.Errorf("failed to get nerdctl data store: %s", err)
		return err
	}

	// create a log URI using nerdctl binary path
	args := map[string]string{
		logging.MagicArgv1: dataStore,
	}
	logURI, err := cio.LogURIGenerator("binary", createOpt.NerdctlCmd, args)
	if err != nil {
		logrus.Errorf("failed to generate a log URI: %s", err)
		return err
	}

	// reset container label with new log URI
	opts, err := cont.Labels(ctx)
	if err != nil {
		logrus.Errorf("failed to get container labels: %s", err)
		return err
	}
	opts[labels.LogURI] = logURI.String()

	// reset OCI spec with new log URI
	spec, err := cont.Spec(ctx)
	if err != nil {
		logrus.Errorf("failed to get container OCI spec: %s", err)
		return err
	}
	spec.Annotations[labels.LogURI] = logURI.String()

	// update container
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
