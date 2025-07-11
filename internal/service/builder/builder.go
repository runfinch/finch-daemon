// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package builder consists of definition of service structures and methods related to build APIs
package builder

import (
	"github.com/runfinch/finch-daemon/api/handlers/builder"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/archive"
	"github.com/runfinch/finch-daemon/pkg/credential"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

type NerdctlService interface {
	backend.NerdctlBuilderSvc
	backend.NerdctlImageSvc
}

type service struct {
	client         backend.ContainerdClient
	nctlBuilderSvc NerdctlService
	logger         flog.Logger
	tarExtractor   archive.TarExtractor
	credentialSvc  *credential.CredentialService
}

// NewService creates a service struct for build APIs.
func NewService(
	client backend.ContainerdClient,
	ncBuilderSvc NerdctlService,
	logger flog.Logger,
	tarExtractor archive.TarExtractor,
	credService *credential.CredentialService,
) builder.Service {
	return &service{
		client:         client,
		nctlBuilderSvc: ncBuilderSvc,
		logger:         logger,
		tarExtractor:   tarExtractor,
		credentialSvc:  credService,
	}
}
