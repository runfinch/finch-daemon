// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"context"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/containerd/nerdctl/v2/pkg/infoutil"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
)

func (s *service) GetInfo(ctx context.Context, config *config.Config) (*dockercompat.Info, error) {
	return infoutil.Info(ctx, s.client.GetClient(), config.Snapshotter, config.CgroupManager)
}
