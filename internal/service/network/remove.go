// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/containerd/nerdctl/pkg/netutil"
	"github.com/runfinch/finch-daemon/internal/service/network/driver"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Remove(ctx context.Context, networkId string) error {
	s.logger.Infof("network delete: network Id %s", networkId)
	net, err := s.getNetwork(networkId)
	if err != nil {
		return fmt.Errorf("failed to find network: %w", err)
	}
	usedNetworkInfo, err := s.netClient.UsedNetworkInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to find used network info: %w", err)
	}
	// API Doc: https://docs.docker.com/engine/api/v1.43/#tag/Network/operation/NetworkDelete
	// does not explicitly call out the scenario when network is in use by a container, although it returns a 403
	if value, ok := usedNetworkInfo[net.Name]; ok {
		return errdefs.NewForbidden(fmt.Errorf("network %q is in use by container %q", networkId, value))
	}

	if net.File == "" {
		return errdefs.NewForbidden(fmt.Errorf("%s is a pre-defined network and cannot be removed", networkId))
	}

	// Ensure thread-safety for network operations using a per-network mutex.
	// RemoveNetwork and CreateNetwork operations on the same network ID are mutually exclusive.
	// Operations on different network IDs can proceed concurrently.
	netMu := s.ensureLock(net.Name)

	netMu.Lock()
	defer netMu.Unlock()

	// Perform additional workflow based on the assigned network labels
	if err := s.handleNetworkLabels(net); err != nil {
		return fmt.Errorf("failed to handle nerdctl label: %w", err)
	}

	err = s.netClient.RemoveNetwork(net)
	if err != nil {
		return fmt.Errorf("failed to remove network: %w", err)
	}

	// Clear the lock if remove was successful
	s.clearLock(net.Name)
	return nil
}

func (s *service) handleNetworkLabels(net *netutil.NetworkConfig) error {
	if net.NerdctlLabels == nil {
		return nil
	}

	for key, value := range *net.NerdctlLabels {
		switch key {
		case driver.FinchICCLabelIPv4:
			if err := s.handleICCLabel(net, value, false); err != nil {
				return fmt.Errorf("error handling IPv4 ICC label: %w", err)
			}
		case driver.FinchICCLabelIPv6:
			if err := s.handleICCLabel(net, value, true); err != nil {
				return fmt.Errorf("error handling IPv6 ICC label: %w", err)
			}
		}
	}
	return nil
}

func (s *service) handleICCLabel(net *netutil.NetworkConfig, value string, isIPv6 bool) error {
	if value != "false" {
		// for some reason the label value got modified.
		// we will still try to remove the iptable rules.
		// iptable.DeleteIfExists is used to ignore non-existent errors
		s.logger.Warnf("unexpected value for ICC label: %s", value)
	}

	bridgeDriver, err := driver.NewBridgeDriver(s.netClient, s.logger, isIPv6)
	if err != nil {
		return fmt.Errorf("unable to create bridge driver: %w", err)
	}

	if err := bridgeDriver.HandleRemove(net); err != nil {
		return fmt.Errorf("error handling ICC label: %w", err)
	}

	return nil
}
