// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/nerdctl/pkg/netutil"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/runfinch/finch-daemon/pkg/utility/maputility"
)

// Create implements the logic to turn a network create request to the back-end nerdctl create network calls.
func (s *service) Create(ctx context.Context, request types.NetworkCreateRequest) (types.NetworkCreateResponse, error) {
	var bridgeDriver BridgeDriverOperations
	var err error

	createOptionsFrom := func(request types.NetworkCreateRequest) (netutil.CreateOptions, error) {
		// Default to "bridge" driver if request does not specify a driver
		driver := request.Driver
		if driver == "" {
			driver = "bridge"
		}

		options := netutil.CreateOptions{
			Name:        request.Name,
			Driver:      driver,
			IPAMDriver:  "default",
			IPAMOptions: request.IPAM.Options,
			Labels:      maputility.Flatten(request.Labels, maputility.KeyEqualsValueFormat),
		}

		if request.IPAM.Driver != "" {
			options.IPAMDriver = request.IPAM.Driver
		}

		if len(request.IPAM.Config) != 0 {
			options.Subnets = []string{}
			if subnet, ok := request.IPAM.Config[0]["Subnet"]; ok {
				options.Subnets = []string{subnet}
			}
			if ipRange, ok := request.IPAM.Config[0]["IPRange"]; ok {
				options.IPRange = ipRange
			}
			if gateway, ok := request.IPAM.Config[0]["Gateway"]; ok {
				options.Gateway = gateway
			}
		}

		// Handle driver-specific options
		switch driver {
		case "bridge":
			bridgeDriver = NewBridgeDriver(s.netClient, s.logger)
			options, err = bridgeDriver.HandleCreateOptions(request, options)
			return options, err
		default:
			options.Options = request.Options
		}

		return options, nil
	}

	if config, err := s.getNetwork(request.Name); err == nil {
		// Network already exists; however, it may not have a network ID.
		response := types.NetworkCreateResponse{
			Warning: fmt.Sprintf("Network with name '%s' already exists", request.Name),
		}
		if config != nil && config.NerdctlID != nil {
			// Share the network ID if it is available.
			response.ID = *config.NerdctlID
			response.Warning = fmt.Sprintf("Network with name '%s' (id: %s) already exists", request.Name, *config.NerdctlID)
		}
		return response, nil
	}

	options, err := createOptionsFrom(request)
	if err != nil {
		return types.NetworkCreateResponse{}, err
	}

	// Create network
	net, err := s.netClient.CreateNetwork(options)
	if err != nil && strings.Contains(err.Error(), "unsupported cni driver") {
		return types.NetworkCreateResponse{}, errdefs.NewNotFound(errPluginNotFound)
	} else if err != nil {
		return types.NetworkCreateResponse{}, err
	} else if net == nil || net.NerdctlID == nil {
		// The create network call to nerdctl was successful, but no network ID was returned.
		// This should not happen.
		return types.NetworkCreateResponse{}, errNetworkIDNotFound
	}

	// Handle post network create actions
	warning := ""
	if bridgeDriver != nil {
		warning, err = bridgeDriver.HandlePostCreate(net)
		if err != nil {
			return types.NetworkCreateResponse{}, err
		}
	}

	return types.NetworkCreateResponse{
		ID:      *net.NerdctlID,
		Warning: warning,
	}, nil
}
