// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/pkg/mocks/mocks_backend"
)

type mockNerdctlService struct {
	*mocks_backend.MockNerdctlBuilderSvc
	*mocks_backend.MockNerdctlImageSvc
}

// TestContainerHandler function is the entry point of container service package's unit test using ginkgo.
func TestContainerService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Build APIs Service")
}
