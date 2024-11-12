// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"

	"github.com/containerd/nerdctl/v2/pkg/infoutil"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
)

//go:generate mockgen --destination=../../mocks/mocks_backend/nerdctlsystemsvc.go -package=mocks_backend github.com/runfinch/finch-daemon/internal/backend NerdctlSystemSvc
type NerdctlSystemSvc interface {
	GetServerVersion(ctx context.Context) (*dockercompat.ServerVersion, error)
}

func (w *NerdctlWrapper) GetServerVersion(ctx context.Context) (*dockercompat.ServerVersion, error) {
	return infoutil.ServerVersion(ctx, w.clientWrapper.GetClient())
}
