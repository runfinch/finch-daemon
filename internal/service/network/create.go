// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/nerdctl/pkg/lockutil"
	"github.com/containerd/nerdctl/pkg/netutil"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/runfinch/finch-daemon/pkg/utility/maputility"
)

// Create implements the logic to turn a network create request to the back-end nerdctl create network calls.
func (s *service) Create(ctx context.Context, request types.NetworkCreateRequest) (types.NetworkCreateResponse, error) {
	// enable_ip_masquerade, host_binding_ipv4, and bridge name network options are not supported by nerdctl.
	// So we must filter out any unsupported options which would prevent the network from being created and accept the defaults.
	bridge := ""
	filterUnsupportedOptions := func(original map[string]string) map[string]string {
		options := map[string]string{}
		for k, v := range original {
			switch k {
			case "com.docker.network.bridge.enable_ip_masquerade":
				// must be true
				if v != "true" {
					s.logger.Warnf("network option com.docker.network.bridge.enable_ip_masquerade is set to %s, but it must be true", v)
				}
			case "com.docker.network.bridge.host_binding_ipv4":
				// must be 0.0.0.0
				if v != "0.0.0.0" {
					s.logger.Warnf("network option com.docker.network.bridge.host_binding_ipv4 is set to %s, but it must be 0.0.0.0", v)
				}
			case "com.docker.network.bridge.name":
				bridge = v
			default:
				options[k] = v
			}
		}
		return options
	}

	createOptionsFrom := func(r types.NetworkCreateRequest) netutil.CreateOptions {
		options := netutil.CreateOptions{
			Name:        r.Name,
			Driver:      "bridge",
			IPAMDriver:  "default",
			IPAMOptions: r.IPAM.Options,
			Options:     filterUnsupportedOptions(r.Options),
			Labels:      maputility.Flatten(r.Labels, maputility.KeyEqualsValueFormat),
		}
		if r.Driver != "" {
			options.Driver = r.Driver
		}
		if r.IPAM.Driver != "" {
			options.IPAMDriver = r.IPAM.Driver
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
		return options
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

	net, err := s.netClient.CreateNetwork(createOptionsFrom(request))
	warning := ""
	if err != nil && strings.Contains(err.Error(), "unsupported cni driver") {
		return types.NetworkCreateResponse{}, errdefs.NewNotFound(errPluginNotFound)
	} else if err != nil {
		return types.NetworkCreateResponse{}, err
	} else if net == nil || net.NerdctlID == nil {
		// The create network call to nerdctl was successful, but no network ID was returned.
		// This should not happen.
		return types.NetworkCreateResponse{}, errNetworkIDNotFound
	}

	// Since nerdctl currently does not support custom bridge names,
	// we explicitly override bridge name in the conflist file for the network that was just created
	if bridge != "" {
		if err = s.setBridgeName(net, bridge); err != nil {
			warning = fmt.Sprintf("Failed to set network bridge name %s: %s", bridge, err)
		}
	}

	return types.NetworkCreateResponse{
		ID:      *net.NerdctlID,
		Warning: warning,
	}, nil
}

// setBridgeName will override the bridge name in an existing CNI config file for a network.
func (s *service) setBridgeName(net *netutil.NetworkConfig, bridge string) error {
	return lockutil.WithDirLock(s.netClient.NetconfPath(), func() error {
		// first, make sure that the bridge name is not used by any of the existing bridge networks
		bridgeNet, err := s.getNetworkByBridgeName(bridge)
		if err != nil {
			return err
		}
		if bridgeNet != nil {
			return fmt.Errorf("bridge name %s already in use by network %s", bridge, bridgeNet.Name)
		}

		// load the CNI config file and set bridge name
		configFilename := s.getConfigPathForNetworkName(net.Name)
		configFile, err := os.Open(configFilename)
		if err != nil {
			return err
		}
		defer configFile.Close()
		var netJSON interface{}
		if err = json.NewDecoder(configFile).Decode(&netJSON); err != nil {
			return err
		}
		netMap, ok := netJSON.(map[string]interface{})
		if !ok {
			return fmt.Errorf("network config file %s is not a valid map", configFilename)
		}
		plugins, ok := netMap["plugins"]
		if !ok {
			return fmt.Errorf("could not find plugins in network config file %s", configFilename)
		}
		pluginsMap, ok := plugins.([]interface{})
		if !ok {
			return fmt.Errorf("could not parse plugins in network config file %s", configFilename)
		}
		for _, plugin := range pluginsMap {
			pluginMap, ok := plugin.(map[string]interface{})
			if !ok {
				continue
			}
			if pluginMap["type"] == "bridge" {
				pluginMap["bridge"] = bridge
				data, err := json.MarshalIndent(netJSON, "", "  ")
				if err != nil {
					return err
				}
				return os.WriteFile(configFilename, data, 0o644)
			}
		}
		return fmt.Errorf("bridge plugin not found in network config file %s", configFilename)
	})
}

// From https://github.com/containerd/nerdctl/blob/v1.5.0/pkg/netutil/netutil.go#L186-L188
func (s *service) getConfigPathForNetworkName(netName string) string {
	return filepath.Join(s.netClient.NetconfPath(), "nerdctl-"+netName+".conflist")
}

type bridgePlugin struct {
	Type   string `json:"type"`
	Bridge string `json:"bridge"`
}

func (s *service) getNetworkByBridgeName(bridge string) (*netutil.NetworkConfig, error) {
	networks, err := s.netClient.FilterNetworks(func(*netutil.NetworkConfig) bool {
		return true
	})
	if err != nil {
		return nil, err
	}
	for _, network := range networks {
		for _, plugin := range network.Plugins {
			if plugin.Network.Type != "bridge" {
				continue
			}
			var bridgeJSON bridgePlugin
			if err = json.Unmarshal(plugin.Bytes, &bridgeJSON); err != nil {
				continue
			}
			if bridgeJSON.Bridge == bridge {
				return network, nil
			}
		}
	}
	return nil, nil
}
