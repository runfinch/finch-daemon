// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/containerd/nerdctl/pkg/lockutil"
	"github.com/containerd/nerdctl/pkg/netutil"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

const (
	FinchICCLabelIPv4           = "finch.network.bridge.enable_icc.ipv4"
	FinchICCLabelIPv6           = "finch.network.bridge.enable_icc.ipv6"
	BridgeICCOption             = "com.docker.network.bridge.enable_icc"
	BridgeHostBindingIpv4Option = "com.docker.network.bridge.host_binding_ipv4"
	BridgeNameOption            = "com.docker.network.bridge.name"
)

type bridgeDriver struct {
	bridgeName string
	disableICC bool
	netClient  backend.NerdctlNetworkSvc
	logger     flog.Logger
	IPv6       bool
}

var _ DriverHandler = (*bridgeDriver)(nil)

var NewBridgeDriver = func(netClient backend.NerdctlNetworkSvc, logger flog.Logger, IPv6 bool) (DriverHandler, error) {
	return &bridgeDriver{
		netClient: netClient,
		logger:    logger,
		IPv6:      IPv6,
	}, nil
}

// HandleCreateOptions processes finch specific options for the bridge driver.
func (bd *bridgeDriver) HandleCreateOptions(request types.NetworkCreateRequest, options netutil.CreateOptions) (netutil.CreateOptions, error) {
	// enable_icc, host_binding_ipv4, and bridge name network options are not supported by nerdctl.
	// So we process these options here and filter them out from the network create request to nerdctl.
	processUnsupportedOptions := func(original map[string]string) map[string]string {
		opts := map[string]string{}
		for k, v := range original {
			switch k {
			case BridgeHostBindingIpv4Option:
				if v != "0.0.0.0" {
					bd.logger.Warnf("network option com.docker.network.bridge.host_binding_ipv4 is set to %s, but it must be 0.0.0.0", v)
				}
			case BridgeICCOption:
				iccOption, err := strconv.ParseBool(v)
				if err != nil {
					bd.logger.Warnf("invalid value for com.docker.network.bridge.enable_icc")
					continue
				}
				bd.disableICC = !iccOption
			case BridgeNameOption:
				bd.bridgeName = v
			default:
				opts[k] = v
			}
		}
		return opts
	}

	options.Options = processUnsupportedOptions(request.Options)

	if bd.disableICC {
		finchICCLabel := FinchICCLabelIPv4
		if bd.IPv6 {
			finchICCLabel = FinchICCLabelIPv6
		}
		options.Labels = append(options.Labels, finchICCLabel+"=false")
	}
	return options, nil
}

func (bd *bridgeDriver) HandlePostCreate(net *netutil.NetworkConfig) (string, error) {
	// Handle bridge driver post create actions
	var warning string
	var err error
	bridgeName := bd.bridgeName
	if bridgeName != "" {
		// Since nerdctl currently does not support custom bridge names,
		// we explicitly override bridge name in the conflist file for the network that was just created
		if err = bd.setBridgeName(net, bridgeName); err != nil {
			warning = fmt.Sprintf("Failed to set network bridge name %s: %s", bridgeName, err)
		}
	}

	if bd.disableICC {
		// Handle "enable_icc=false" option if set (bd.disableICC is true)
		// By default, CNI allows connectivity between containers attached to the same bridge.
		// If "com.docker.network.bridge.enable_icc" option is explicitly set to false,
		// we disable inter-container connectivity by applying iptable rules
		// If "com.docker.network.bridge.enable_icc=true" is set, it is considered a noop
		if bridgeName == "" {
			bridgeName, err = bd.getBridgeName(net)
			if err != nil {
				return "", fmt.Errorf("failed to get bridge name to enable inter-container connectivity: %w ", err)
			}
		}

		err = bd.addICCDropRule(bridgeName)
		if err != nil {
			return "", fmt.Errorf("failed to disable inter-container connectivity: %w", err)
		}
	}

	return warning, nil
}

func (bd *bridgeDriver) HandleRemove(net *netutil.NetworkConfig) error {
	bridgeName, err := bd.getBridgeName(net)
	if err != nil {
		return fmt.Errorf("failed to get bridge name to remove inter-container connectivity: %w ", err)
	}
	err = bd.removeICCDropRule(bridgeName)
	if err != nil {
		return fmt.Errorf("failed to remove iptables DROP rule : %w", err)
	}
	return nil
}

// setBridgeName will override the bridge name in an existing CNI config file for a network.
func (bd *bridgeDriver) setBridgeName(net *netutil.NetworkConfig, bridgeName string) error {
	return lockutil.WithDirLock(bd.netClient.NetconfPath(), func() error {
		// first, make sure that the bridge name is not used by any of the existing bridge networks
		bridgeNet, err := bd.getNetworkByBridgeName(bridgeName)
		if err != nil {
			return err
		}
		if bridgeNet != nil {
			return fmt.Errorf("bridge name %s already in use by network %s", bridgeName, bridgeNet.Name)
		}

		configFilename := bd.getConfigPathForNetworkName(net.Name)
		netMap, bridgePlugin, err := bd.parseBridgeConfig(configFilename)
		if err != nil {
			return err
		}

		// Update the bridge name in the bridge plugin
		bridgePlugin["bridge"] = bridgeName

		// Update the plugins in the full config
		plugins := netMap["plugins"].([]interface{})
		for i, plugin := range plugins {
			if p, ok := plugin.(map[string]interface{}); ok && p["type"] == "bridge" {
				plugins[i] = bridgePlugin
				break
			}
		}

		data, err := json.MarshalIndent(netMap, "", "  ")
		if err != nil {
			return err
		}

		// Write the updated config back to the file with the original permissions
		fileInfo, err := os.Stat(configFilename)
		if err != nil {
			return err
		}

		return os.WriteFile(configFilename, data, fileInfo.Mode().Perm())
	})
}

func (bd *bridgeDriver) getBridgeName(net *netutil.NetworkConfig) (string, error) {
	var bridgeName string
	err := lockutil.WithDirLock(bd.netClient.NetconfPath(), func() error {
		configFilename := bd.getConfigPathForNetworkName(net.Name)
		_, bridgePlugin, err := bd.parseBridgeConfig(configFilename)
		if err != nil {
			return err
		}

		bridge, ok := bridgePlugin["bridge"].(string)
		if !ok {
			return fmt.Errorf("bridge name in config file %s is not a string", configFilename)
		}
		bridgeName = bridge
		return nil
	})

	if err != nil {
		return "", err
	}

	return bridgeName, nil
}

func (bd *bridgeDriver) parseBridgeConfig(configFilename string) (map[string]interface{}, map[string]interface{}, error) {
	configFile, err := os.Open(configFilename)
	if err != nil {
		return nil, nil, err
	}
	defer configFile.Close()

	var netJSON interface{}
	if err = json.NewDecoder(configFile).Decode(&netJSON); err != nil {
		return nil, nil, err
	}

	netMap, ok := netJSON.(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("network config file %s is not a valid map", configFilename)
	}

	plugins, ok := netMap["plugins"]
	if !ok {
		return nil, nil, fmt.Errorf("could not find plugins in network config file %s", configFilename)
	}

	pluginsMap, ok := plugins.([]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("could not parse plugins in network config file %s", configFilename)
	}

	for _, plugin := range pluginsMap {
		pluginMap, ok := plugin.(map[string]interface{})
		if !ok {
			continue
		}
		if pluginMap["type"] == "bridge" {
			return netMap, pluginMap, nil
		}
	}

	return nil, nil, fmt.Errorf("bridge plugin not found in network config file %s", configFilename)
}

func (bd *bridgeDriver) getNetworkByBridgeName(bridgeName string) (*netutil.NetworkConfig, error) {
	networks, err := bd.netClient.FilterNetworks(func(*netutil.NetworkConfig) bool {
		return true
	})
	if err != nil {
		return nil, err
	}

	var bridgePlugin struct {
		Type   string `json:"type"`
		Bridge string `json:"bridge"`
	}

	for _, network := range networks {
		for _, plugin := range network.Plugins {
			if plugin.Network.Type != "bridge" {
				continue
			}

			if err = json.Unmarshal(plugin.Bytes, &bridgePlugin); err != nil {
				continue
			}
			if bridgePlugin.Bridge == bridgeName {
				return network, nil
			}
		}
	}
	return nil, nil
}

func (bd *bridgeDriver) addICCDropRule(bridgeIface string) error {
	bd.logger.Debugf("adding ICC drop rule for bridge: %s", bridgeIface)
	iccDropRule := []string{"-i", bridgeIface, "-o", bridgeIface, "-j", "DROP"}
	ipc, err := newIptablesCommand(bd.IPv6)
	if err != nil {
		return err
	}

	err = ipc.AddRule(iccDropRule...)
	if err != nil {
		return fmt.Errorf("failed to add iptables rule to drop ICC: %v", err)
	}

	return nil
}

func (bd *bridgeDriver) removeICCDropRule(bridgeIface string) error {
	bd.logger.Debugf("removing ICC drop rule for bridge: %s", bridgeIface)
	iccDropRule := []string{"-i", bridgeIface, "-o", bridgeIface, "-j", "DROP"}
	ipc, err := newIptablesCommand(bd.IPv6)
	if err != nil {
		return err
	}

	err = ipc.DelRule(iccDropRule...)
	if err != nil {
		return fmt.Errorf("failed to remove iptables rules to drop ICC: %v", err)
	}

	return nil
}

// From https://github.com/containerd/nerdctl/blob/v1.5.0/pkg/netutil/netutil.go#L186-L188
func (bd *bridgeDriver) getConfigPathForNetworkName(netName string) string {
	return filepath.Join(bd.netClient.NetconfPath(), "nerdctl-"+netName+".conflist")
}
