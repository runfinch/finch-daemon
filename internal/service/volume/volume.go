// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// package volume defines the volumes service.
package volume

import (
	"github.com/runfinch/finch-daemon/api/handlers/volume"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

type service struct {
	nctlVolumeSvc backend.NerdctlVolumeSvc
	logger        flog.Logger
}

func NewService(nerdctlVolumeSvc backend.NerdctlVolumeSvc, logger flog.Logger) volume.Service {
	return &service{
		nctlVolumeSvc: nerdctlVolumeSvc,
		logger:        logger,
	}
}
