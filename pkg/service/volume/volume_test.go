// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestVolumeService function is the entry point of volume service package's unit test using ginkgo
func TestVolumeService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Volumes APIs Service")
}
