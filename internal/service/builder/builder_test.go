// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/api/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
)

type mockNerdctlService struct {
	*mocks_backend.MockNerdctlBuilderSvc
	*mocks_backend.MockNerdctlImageSvc
}

// Build implements the Build method from NerdctlService interface with the buildID parameter.
func (m mockNerdctlService) Build(ctx context.Context, client backend.ContainerdClient, options types.BuilderBuildOptions, buildID string) error {
	return m.MockNerdctlBuilderSvc.Build(ctx, client, options, buildID)
}

// TestContainerHandler function is the entry point of container service package's unit test using ginkgo.
func TestContainerService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Build APIs Service")
}
