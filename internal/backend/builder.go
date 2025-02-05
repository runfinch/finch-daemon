// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"

	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/buildkitutil"
	"github.com/containerd/nerdctl/v2/pkg/cmd/builder"
)

//go:generate mockgen --destination=../../mocks/mocks_backend/nerdctlbuildersvc.go -package=mocks_backend github.com/runfinch/finch-daemon/internal/backend NerdctlBuilderSvc
type NerdctlBuilderSvc interface {
	Build(ctx context.Context, client ContainerdClient, options types.BuilderBuildOptions) error
	GetBuildkitHost() (string, error)
}

func (*NerdctlWrapper) Build(ctx context.Context, client ContainerdClient, options types.BuilderBuildOptions) error {
	return builder.Build(ctx, client.GetClient(), options)
}

func (w *NerdctlWrapper) GetBuildkitHost() (string, error) {
	return buildkitutil.GetBuildkitHost(w.globalOptions.Namespace)
}
