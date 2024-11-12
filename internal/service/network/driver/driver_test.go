// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestNetworkService is the entry point of the network service package's unit tests using ginkgo.
func TestNetworkService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Network Driver Operations")
}
