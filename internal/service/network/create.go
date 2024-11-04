// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/nerdctl/pkg/netutil"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/internal/service/network/driver"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/runfinch/finch-daemon/pkg/utility/maputility"
)

// Create implements the logic to turn a network create request to the back-end nerdctl create network calls.
func (s *service) Create(ctx context.Context, request types.NetworkCreateRequest) (types.NetworkCreateResponse, error) {
	var bridgeDriver driver.DriverHandler
	var err error

	createOptionsFrom := func(request types.NetworkCreateRequest) (netutil.CreateOptions, error) {
		// Default to "bridge" driver if request does not specify a driver
		networkDriver := request.Driver
		if networkDriver == "" {
			networkDriver = "bridge"
		}

		options := netutil.CreateOptions{
			Name:        request.Name,
			Driver:      networkDriver,
			IPAMDriver:  "default",
			IPAMOptions: request.IPAM.Options,
			Labels:      maputility.Flatten(request.Labels, maputility.KeyEqualsValueFormat),
			IPv6:        request.EnableIPv6,
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
		switch networkDriver {
		case "bridge":
			bridgeDriver, err = driver.NewBridgeDriver(s.netClient, s.logger, options.IPv6)
			if err != nil {
				return options, err
			}
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

	// Ensure thread-safety for network operations using a per-network mutex.
	// Operations on different network IDs can proceed concurrently.
	netMu := s.ensureLock(request.Name)

	netMu.Lock()
	defer netMu.Unlock()

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

	// Add cleanup func to remove the network if an error is encountered during post processing
	cleanup := func(ctx context.Context, name string) {
		if cleanupErr := s.Remove(ctx, name); cleanupErr != nil {
			// ignore if the network does not exist
			if !errdefs.IsNotFound(cleanupErr) {
				s.logger.Errorf("cleanup failed in defer %s: %v", name, cleanupErr)
			}
		}
	}

	defer func() {
		if err != nil {
			cleanup(ctx, request.Name)
		}
	}()

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
