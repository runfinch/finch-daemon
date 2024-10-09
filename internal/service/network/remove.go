// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/containerd/nerdctl/pkg/netutil"
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

	// Perform additional workflow based on the assigned network labels
	if err := s.handleNetworkLabels(net); err != nil {
		return fmt.Errorf("failed to handle nerdctl label: %w", err)
	}

	return s.netClient.RemoveNetwork(net)
}

func (s *service) handleNetworkLabels(net *netutil.NetworkConfig) error {
	if net.NerdctlLabels == nil {
		return nil
	}

	for key, value := range *net.NerdctlLabels {
		switch key {
		case FinchICCLabel:
			if err := s.handleEnableICCOption(net, value); err != nil {
				return fmt.Errorf("error handling %s label: %w", BridgeICCOption, err)
			}
		}
	}
	return nil
}

func (s *service) handleEnableICCOption(net *netutil.NetworkConfig, value string) error {
	if value != "false" {
		// for some reason the label value got modified.
		// we will still try to remove the iptable rules.
		// iptable.DeleteIfExists is used to ignore non-existent errors
		s.logger.Warnf("unexpected value for %s label: %s", BridgeICCOption, value)
	}
	// Remove iptable rules set for disabling ICC for the network bridge
	bridgeDriver := NewBridgeDriver(s.netClient, s.logger)
	bridgeName, err := bridgeDriver.GetBridgeName(net)
	if err != nil {
		return fmt.Errorf("unable to get bridge name: %w", err)
	}
	err = bridgeDriver.DisableICC(bridgeName, false)
	if err != nil {
		return fmt.Errorf("unable to remove ICC disable rule: %w", err)
	}
	return nil
}
