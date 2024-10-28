// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"errors"
	"fmt"

	"github.com/coreos/go-iptables/iptables"
)

const (
	statusChainExists = 1

	FilterTableName   = "filter"
	ForwardChainName  = "FORWARD"
	FinchIsolateChain = "FINCH-ISOLATE-CHAIN"
)

// IPTablesWrapper is an interface that wraps the methods of iptables.IPTables
// to help with mock
//
//go:generate mockgen --destination=../../../../mocks/mocks_network/iptables.go -package=mocks_network  . IPTablesWrapper
type IPTablesWrapper interface {
	ChainExists(table, chain string) (bool, error)
	NewChain(table, chain string) error
	InsertUnique(table, chain string, pos int, rulespec ...string) error
	AppendUnique(table, chain string, rulespec ...string) error
	DeleteIfExists(table, chain string, rulespec ...string) error
	DeleteChain(table, chain string) error
}

type iptablesCommand struct {
	protos map[iptables.Protocol]IPTablesWrapper
}

var newIptablesCommand = func(ipv6 bool) (*iptablesCommand, error) {
	iptCommand := &iptablesCommand{
		protos: make(map[iptables.Protocol]IPTablesWrapper),
	}

	protocols := []iptables.Protocol{iptables.ProtocolIPv4}
	if ipv6 {
		protocols = append(protocols, iptables.ProtocolIPv6)
	}

	for _, proto := range protocols {
		ipt, err := iptables.NewWithProtocol(proto)
		if err != nil {
			return nil, fmt.Errorf("could not initialize iptables protocol %v: %w", proto, err)
		}
		iptCommand.protos[proto] = ipt
	}

	return iptCommand, nil
}

// EnsureChain creates a new iptables chain if one doesn't exist already.
func (ipc *iptablesCommand) ensureChain(ipt IPTablesWrapper, table, chain string) error {
	if ipt == nil {
		return errors.New("failed to ensure iptable chain: IPTables was nil")
	}
	exists, err := ipt.ChainExists(table, chain)
	if err != nil {
		return fmt.Errorf("failed to check if iptables %s exists: %v", chain, err)
	}
	if !exists {
		err = ipt.NewChain(table, chain)
		if err != nil {
			eerr, eok := err.(*iptables.Error)
			if !eok {
				// type assertion failed, return the original error
				return fmt.Errorf("failed to create %s chain: %w", chain, err)
			}

			// ignore if the chain was created in the meantime
			if eerr.ExitStatus() != statusChainExists {
				return fmt.Errorf("failed to create %s chain: %w", chain, err)
			}
		}
	}
	return nil
}

func (ipc *iptablesCommand) cleanupChain(ipt IPTablesWrapper, tableName string, chainName string) {
	// attempt to delete the chain if it exists and is empty
	ipt.DeleteChain(tableName, chainName)
}

func (ipc *iptablesCommand) setupChains(ipt IPTablesWrapper) error {
	if err := ipc.ensureChain(ipt, FilterTableName, FinchIsolateChain); err != nil {
		ipc.cleanupChain(ipt, FilterTableName, FinchIsolateChain)
		return err
	}
	// Add a rule to the FORWARD chain that jumps to the FINCH-ISOLATE-CHAIN for all packets
	jumpRule := []string{"-j", FinchIsolateChain}
	if err := ipt.InsertUnique(FilterTableName, ForwardChainName, 1, jumpRule...); err != nil {
		return fmt.Errorf("failed to add jump rule to FINCH-ISOLATE-CHAIN chain: %v", err)
	}
	// In the FINCH-ISOLATE-CHAIN, add a rule to return to the FORWARD chain when no match
	returnRule := []string{"-j", "RETURN"}
	if err := ipt.AppendUnique(FilterTableName, FinchIsolateChain, returnRule...); err != nil {
		return fmt.Errorf("failed to add RETURN rule in FINCH-ISOLATE-CHAIN chain: %v", err)
	}
	return nil
}

func (ipc *iptablesCommand) cleanupRules(ipt IPTablesWrapper, tableName string, chainName string, rulespec ...string) {
	ipt.DeleteIfExists(tableName, chainName, rulespec...)
}

func (ipc *iptablesCommand) addRule(ipt IPTablesWrapper, rulespec ...string) error {
	if err := ipc.setupChains(ipt); err != nil {
		return err
	}

	if err := ipt.InsertUnique(FilterTableName, FinchIsolateChain, 1, rulespec...); err != nil {
		ipc.cleanupRules(ipt, FilterTableName, FinchIsolateChain, rulespec...)
		return fmt.Errorf("failed to add iptables rule: %w", err)
	}

	return nil
}

func (ipc *iptablesCommand) AddRule(rulespec ...string) error {
	for _, ipt := range ipc.protos {
		if err := ipc.addRule(ipt, rulespec...); err != nil {
			return err
		}
	}
	return nil
}

func (ipc *iptablesCommand) delRule(ipt IPTablesWrapper, rulespec ...string) error {
	return ipt.DeleteIfExists(FilterTableName, FinchIsolateChain, rulespec...)
}

func (ipc *iptablesCommand) DelRule(rulespec ...string) error {
	for _, ipt := range ipc.protos {
		if err := ipc.delRule(ipt, rulespec...); err != nil {
			return err
		}
	}
	return nil
}
