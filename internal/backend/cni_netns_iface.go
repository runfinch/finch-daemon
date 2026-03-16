// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package backend provides the interface and implementations for CNI + netns
// lifecycle operations used by Solution IV (hook-free networking).
package backend

import "context"

// ContainerCNISvc manages the pre-created network namespace and CNI lifecycle
// for a container, replacing the OCI hook mechanism.
//
//go:generate mockgen --destination=../../mocks/mocks_backend/containercnisvc.go -package=mocks_backend github.com/runfinch/finch-daemon/internal/backend ContainerCNISvc
type ContainerCNISvc interface {
	// SetupContainerNetwork creates a named netns, runs CNI Setup against it,
	// and returns the netns path.  The caller must call RemoveContainerNetwork
	// on cleanup.
	SetupContainerNetwork(ctx context.Context, containerID, networkName string) (netnsPath string, err error)

	// RemoveContainerNetwork tears down the CNI network and deletes the netns.
	RemoveContainerNetwork(ctx context.Context, containerID, networkName, netnsPath string) error
}
