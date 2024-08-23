// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"github.com/runfinch/finch-daemon/api/handlers/system"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

type service struct {
	client      backend.ContainerdClient
	ncSystemSvc backend.NerdctlSystemSvc
	logger      flog.Logger
}

func NewService(client backend.ContainerdClient, ncSystemSvc backend.NerdctlSystemSvc, logger flog.Logger) system.Service {
	return &service{
		client:      client,
		logger:      logger,
		ncSystemSvc: ncSystemSvc,
	}
}
