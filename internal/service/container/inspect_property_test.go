// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"math/rand"
	"testing"
	"testing/quick"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
)

// TestGetHostConfigFromDockerCompat_NeverPanics verifies that getHostConfigFromDockerCompat
// never panics for any valid input, including nil, empty, and randomly populated structs.
// Property: for all inputs, the function either returns nil (for nil input) or a non-nil result.
func TestGetHostConfigFromDockerCompat_NeverPanics(t *testing.T) {
	// Property 1: nil input always returns nil
	if result := getHostConfigFromDockerCompat(nil); result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}

	// Property 2: non-nil input always returns non-nil
	f := func(
		cgroupnsMode string,
		dns []string,
		dnsOptions []string,
		dnsSearch []string,
		extraHosts []string,
		groupAdd []string,
		utsMode string,
		runtime string,
		cpuShares uint64,
		cpuPeriod uint64,
		cpuQuota int64,
		memory int64,
		memorySwap int64,
		blkioWeight uint16,
		readonlyRootfs bool,
		oomKillDisable bool,
	) bool {
		input := &dockercompat.HostConfig{
			CgroupnsMode:   cgroupnsMode,
			DNS:            dns,
			DNSOptions:     dnsOptions,
			DNSSearch:      dnsSearch,
			ExtraHosts:     extraHosts,
			GroupAdd:       groupAdd,
			UTSMode:        utsMode,
			Runtime:        runtime,
			CPUShares:      cpuShares,
			CPUPeriod:      cpuPeriod,
			CPUQuota:       cpuQuota,
			Memory:         memory,
			MemorySwap:     memorySwap,
			BlkioSettings:  dockercompat.BlkioSettings{BlkioWeight: blkioWeight},
			ReadonlyRootfs: readonlyRootfs,
			OomKillDisable: oomKillDisable,
		}
		result := getHostConfigFromDockerCompat(input)
		if result == nil {
			return false
		}
		// Verify scalar fields are preserved exactly
		if string(result.CgroupnsMode) != cgroupnsMode {
			return false
		}
		if result.UTSMode != utsMode {
			return false
		}
		if result.Runtime != runtime {
			return false
		}
		if result.CPUShares != int64(cpuShares) {
			return false
		}
		if result.CPUPeriod != int64(cpuPeriod) {
			return false
		}
		if result.CPUQuota != cpuQuota {
			return false
		}
		if result.Memory != memory {
			return false
		}
		if result.MemorySwap != memorySwap {
			return false
		}
		if result.BlkioWeight != blkioWeight {
			return false
		}
		if result.ReadonlyRootfs != readonlyRootfs {
			return false
		}
		if result.OomKillDisable != oomKillDisable {
			return false
		}
		return true
	}

	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(42)), //nolint:gosec // deterministic seed for reproducibility
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("property violated: %v", err)
	}
}

// TestGetHostConfigFromDockerCompat_SlicePreservation verifies that slice fields
// (DNS, DNSOptions, DNSSearch, ExtraHosts, GroupAdd) are preserved exactly.
// Property: output slice length == input slice length for all non-nil slices.
func TestGetHostConfigFromDockerCompat_SlicePreservation(t *testing.T) {
	f := func(dns []string, dnsOptions []string, dnsSearch []string, extraHosts []string, groupAdd []string) bool {
		input := &dockercompat.HostConfig{
			DNS:        dns,
			DNSOptions: dnsOptions,
			DNSSearch:  dnsSearch,
			ExtraHosts: extraHosts,
			GroupAdd:   groupAdd,
		}
		result := getHostConfigFromDockerCompat(input)
		if result == nil {
			return false
		}
		if len(result.DNS) != len(dns) {
			return false
		}
		if len(result.DNSOptions) != len(dnsOptions) {
			return false
		}
		if len(result.DNSSearch) != len(dnsSearch) {
			return false
		}
		if len(result.ExtraHosts) != len(extraHosts) {
			return false
		}
		if len(result.GroupAdd) != len(groupAdd) {
			return false
		}
		return true
	}

	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(99)), //nolint:gosec // deterministic seed for reproducibility
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("slice preservation property violated: %v", err)
	}
}

// TestGetHostConfigFromDockerCompat_BlkioDeviceFiltersNil verifies that nil entries
// in blkio device slices are always filtered out.
// Property: output device count <= input device count (nils are dropped).
func TestGetHostConfigFromDockerCompat_BlkioDeviceFiltersNil(t *testing.T) {
	// Generate inputs with a mix of nil and non-nil entries
	type blkioInput struct {
		nilCount    int
		nonNilCount int
	}

	cases := []blkioInput{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
		{3, 2},
		{0, 5},
		{5, 0},
	}

	for _, tc := range cases {
		weightDevices := make([]*dockercompat.WeightDevice, tc.nilCount+tc.nonNilCount)
		for i := 0; i < tc.nonNilCount; i++ {
			weightDevices[i] = &dockercompat.WeightDevice{Path: "/dev/sda", Weight: 100}
		}
		// remaining entries are nil by default

		throttleDevices := make([]*dockercompat.ThrottleDevice, tc.nilCount+tc.nonNilCount)
		for i := 0; i < tc.nonNilCount; i++ {
			throttleDevices[i] = &dockercompat.ThrottleDevice{Path: "/dev/sda", Rate: 1024}
		}

		input := &dockercompat.HostConfig{
			BlkioSettings: dockercompat.BlkioSettings{
				BlkioWeightDevice:    weightDevices,
				BlkioDeviceReadBps:   throttleDevices,
				BlkioDeviceWriteBps:  throttleDevices,
				BlkioDeviceReadIOps:  throttleDevices,
				BlkioDeviceWriteIOps: throttleDevices,
			},
		}
		result := getHostConfigFromDockerCompat(input)
		if result == nil {
			t.Errorf("expected non-nil result")
			continue
		}
		if len(result.BlkioWeightDevice) != tc.nonNilCount {
			t.Errorf("BlkioWeightDevice: expected %d non-nil entries, got %d", tc.nonNilCount, len(result.BlkioWeightDevice))
		}
		if len(result.BlkioDeviceReadBps) != tc.nonNilCount {
			t.Errorf("BlkioDeviceReadBps: expected %d non-nil entries, got %d", tc.nonNilCount, len(result.BlkioDeviceReadBps))
		}
	}
}
