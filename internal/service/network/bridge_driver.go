// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/containerd/nerdctl/pkg/lockutil"
	"github.com/containerd/nerdctl/pkg/netutil"
	"github.com/coreos/go-iptables/iptables"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

const (
	FinchICCLabel = "finch.network.bridge.enable_icc"

	BridgeICCOption             = "com.docker.network.bridge.enable_icc"
	BridgeHostBindingIpv4Option = "com.docker.network.bridge.host_binding_ipv4"
	BridgeNameOption            = "com.docker.network.bridge.name"
)

//go:generate mockgen --destination=../../../mocks/mocks_network/bridge_driver.go -package=mocks_network -mock_names BridgeDriverOperations=BridgeDriver . BridgeDriverOperations
type BridgeDriverOperations interface {
	HandleCreateOptions(request types.NetworkCreateRequest, options netutil.CreateOptions) (netutil.CreateOptions, error)
	HandlePostCreate(net *netutil.NetworkConfig) (string, error)
	SetBridgeName(net *netutil.NetworkConfig, bridge string) error
	GetBridgeName(net *netutil.NetworkConfig) (string, error)
	GetNetworkByBridgeName(bridge string) (*netutil.NetworkConfig, error)
	DisableICC(bridgeIface string, insert bool) error
	SetICCDisabled()
	ICCDisabled() bool
}

// IPTablesWrapper is an interface that wraps the methods of iptables.IPTables
// to help with mock
//
//go:generate mockgen --destination=../../../mocks/mocks_network/iptables.go -package=mocks_network  . IPTablesWrapper
type IPTablesWrapper interface {
	ChainExists(table, chain string) (bool, error)
	NewChain(table, chain string) error
	Insert(table, chain string, pos int, rulespec ...string) error
	Append(table, chain string, rulespec ...string) error
	DeleteIfExists(table, chain string, rulespec ...string) error
}

// IPTablesWrapperImpl implements IPTablesWrapper
// that delegates to an actual iptables.IPTables instance.
type IPTablesWrapperImpl struct {
	ipt *iptables.IPTables
}

func NewIPTablesWrapper() IPTablesWrapper {
	iptables, _ := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	return &IPTablesWrapperImpl{ipt: iptables}
}

func (i *IPTablesWrapperImpl) ChainExists(table, chain string) (bool, error) {
	return i.ipt.ChainExists(table, chain)
}

func (i *IPTablesWrapperImpl) NewChain(table, chain string) error {
	return i.ipt.NewChain(table, chain)
}

func (i *IPTablesWrapperImpl) Insert(table, chain string, pos int, rulespec ...string) error {
	return i.ipt.Insert(table, chain, pos, rulespec...)
}

func (i *IPTablesWrapperImpl) Append(table, chain string, rulespec ...string) error {
	return i.ipt.Append(table, chain, rulespec...)
}

func (i *IPTablesWrapperImpl) DeleteIfExists(table, chain string, rulespec ...string) error {
	return i.ipt.DeleteIfExists(table, chain, rulespec...)
}

type bridgeDriver struct {
	bridgeName string
	disableICC bool
	netClient  backend.NerdctlNetworkSvc
	logger     flog.Logger
	ipt        IPTablesWrapper
}

var _ BridgeDriverOperations = (*bridgeDriver)(nil)

var NewBridgeDriver = func(netClient backend.NerdctlNetworkSvc, logger flog.Logger) BridgeDriverOperations {
	return &bridgeDriver{
		netClient: netClient,
		logger:    logger,
		ipt:       NewIPTablesWrapper(),
	}
}

// handleBridgeDriverOptions filters unsupported options for the bridge driver.
func (bd *bridgeDriver) HandleCreateOptions(request types.NetworkCreateRequest, options netutil.CreateOptions) (netutil.CreateOptions, error) {
	// enable_icc, host_binding_ipv4, and bridge name network options are not supported by nerdctl.
	// So we must filter out any unsupported options which would prevent the network from being created and accept the defaults.
	filterUnsupportedOptions := func(original map[string]string) map[string]string {
		opts := map[string]string{}
		for k, v := range original {
			switch k {
			case BridgeHostBindingIpv4Option:
				if v != "0.0.0.0" {
					bd.logger.Warnf("network option com.docker.network.bridge.host_binding_ipv4 is set to %s, but it must be 0.0.0.0", v)
				}
			case BridgeICCOption:
				iccOption, err := strconv.ParseBool(v) //t
				if err != nil {
					bd.logger.Warnf("invalid value for com.docker.network.bridge.enable_icc: %s", v)
				}
				if !iccOption {
					bd.SetICCDisabled()
				}

			case BridgeNameOption:
				bd.bridgeName = v
			default:
				opts[k] = v
			}
		}
		return opts
	}

	options.Options = filterUnsupportedOptions(request.Options)

	if bd.ICCDisabled() {
		// Append a label when enable_icc is set to false. This is used for clean up during network removal.
		options.Labels = append(options.Labels, FinchICCLabel+"=false")
	}

	// Return the modified options
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
		if err = bd.SetBridgeName(net, bridgeName); err != nil {
			warning = fmt.Sprintf("Failed to set network bridge name %s: %s", bridgeName, err)
		}
	}

	if bd.ICCDisabled() {
		// Handle "enable_icc=false" option if set (bd.disableICC is true)
		// By default, CNI allows connectivity between containers attached to the same bridge.
		// If "com.docker.network.bridge.enable_icc" option is explicitly set to false,
		// we disable inter-container connectivity by applying iptable rules
		// If "com.docker.network.bridge.enable_icc=true" is set, it is considered a noop
		if bridgeName == "" {
			bridgeName, err = bd.GetBridgeName(net)
			if err != nil {
				return "", fmt.Errorf("failed to get bridge name to enable inter-container connectivity: %w ", err)
			}
		}
		err = bd.DisableICC(bridgeName, true)
		if err != nil {
			return "", fmt.Errorf("failed to disable inter-container connectivity: %w", err)
		}
	}

	return warning, nil
}

// setBridgeName will override the bridge name in an existing CNI config file for a network.
func (bd *bridgeDriver) SetBridgeName(net *netutil.NetworkConfig, bridge string) error {
	return lockutil.WithDirLock(bd.netClient.NetconfPath(), func() error {
		// first, make sure that the bridge name is not used by any of the existing bridge networks
		bridgeNet, err := bd.GetNetworkByBridgeName(bridge)
		if err != nil {
			return err
		}
		if bridgeNet != nil {
			return fmt.Errorf("bridge name %s already in use by network %s", bridge, bridgeNet.Name)
		}

		// load the CNI config file and set bridge name
		configFilename := bd.getConfigPathForNetworkName(net.Name)
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

func (bd *bridgeDriver) GetBridgeName(net *netutil.NetworkConfig) (string, error) {
	var bridgeName string
	err := lockutil.WithDirLock(bd.netClient.NetconfPath(), func() error {
		configFilename := bd.getConfigPathForNetworkName(net.Name)
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
				bridge, ok := pluginMap["bridge"].(string)
				if !ok {
					return fmt.Errorf("bridge name in config file %s is not a string", configFilename)
				}
				bridgeName = bridge
				return nil
			}
		}

		return fmt.Errorf("bridge plugin not found in network config file %s", configFilename)
	})

	if err != nil {
		return "", err
	}

	return bridgeName, nil
}

type bridgePlugin struct {
	Type   string `json:"type"`
	Bridge string `json:"bridge"`
}

func (bd *bridgeDriver) GetNetworkByBridgeName(bridge string) (*netutil.NetworkConfig, error) {
	networks, err := bd.netClient.FilterNetworks(func(*netutil.NetworkConfig) bool {
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

func (bd *bridgeDriver) SetICCDisabled() {
	bd.disableICC = true
}

func (bd *bridgeDriver) ICCDisabled() bool {
	return bd.disableICC
}

func (bd *bridgeDriver) DisableICC(bridgeIface string, insert bool) error {
	filterTable := "filter"
	isolateChain := "FINCH-ISOLATE-CHAIN"

	if bd.ipt == nil {
		return fmt.Errorf("iptables is not initialized")
	}

	// Check if the FINCH-ISOLATE-CHAIN already exists
	exists, err := bd.ipt.ChainExists(filterTable, isolateChain)
	if err != nil {
		return fmt.Errorf("failed to check if %s chain exists: %v", isolateChain, err)
	}

	if !exists {
		// Create and setup the FINCH-ISOLATE-CHAIN chain if it doesn't exist
		err = bd.ipt.NewChain(filterTable, isolateChain)
		if err != nil {
			return fmt.Errorf("failed to create %s chain: %v", isolateChain, err)
		}
		// Add a rule to the FORWARD chain that jumps to the FINCH-ISOLATE-CHAIN for all packets
		jumpRule := []string{"-j", isolateChain}
		err = bd.ipt.Insert(filterTable, "FORWARD", 1, jumpRule...)
		if err != nil {
			return fmt.Errorf("failed to add %s jump rule to FORWARD chain: %v", isolateChain, err)
		}
		// In the FINCH-ISOLATE-CHAIN, add a rule to return to the FORWARD chain if no match
		returnRule := []string{"-j", "RETURN"}
		err = bd.ipt.Append(filterTable, isolateChain, returnRule...)
		if err != nil {
			return fmt.Errorf("failed to add RETURN rule to DOCKER-ISOLATE chain: %v", err)
		}
	}

	// In the FINCH-ISOLATE-CHAIN, add or remove the DROP rule for packets from and to the same bridge
	dropRule := []string{"-i", bridgeIface, "-o", bridgeIface, "-j", "DROP"}
	if insert {
		err = bd.ipt.Insert(filterTable, isolateChain, 1, dropRule...)
		if err != nil {
			return fmt.Errorf("failed to add DROP rule to %s chain: %v", isolateChain, err)
		}
	} else {
		err = bd.ipt.DeleteIfExists(filterTable, isolateChain, dropRule...)
		if err != nil {
			return fmt.Errorf("failed to remove DROP rule from %s chain: %v", isolateChain, err)
		}
	}
	return nil
}

// From https://github.com/containerd/nerdctl/blob/v1.5.0/pkg/netutil/netutil.go#L186-L188
func (bd *bridgeDriver) getConfigPathForNetworkName(netName string) string {
	return filepath.Join(bd.netClient.NetconfPath(), "nerdctl-"+netName+".conflist")
}
