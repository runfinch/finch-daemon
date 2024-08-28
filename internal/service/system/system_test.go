// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestSystemService is the entry point of system service package's unit tests using ginkgo.
func TestSystemService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - System APIs Service")
}
