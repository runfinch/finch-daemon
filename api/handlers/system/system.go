// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package system contains functions and structures related to system level APIs
package system

import (
	"context"
	"net/http"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"

	eventtype "github.com/runfinch/finch-daemon/api/events"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

// Service defines the interface for functions that provide supplementary information to the handlers
//
//go:generate mockgen --destination=../../../mocks/mocks_system/systemsvc.go -package=mocks_system github.com/runfinch/finch-daemon/api/handlers/system Service
type Service interface {
	Auth(ctx context.Context, username, password, serverAddr string) (string, error)
	SubscribeEvents(ctx context.Context, filters map[string][]string) (<-chan *eventtype.Event, <-chan error)
	GetInfo(ctx context.Context, config *config.Config) (*dockercompat.Info, error)
	GetVersion(ctx context.Context) (*types.VersionInfo, error)
}

// RegisterHandlers sets up the handlers and assigns the handlers to the router. Both r and versionedR are used
// as `GET /version` is called by the docker python client without the API version number.
func RegisterHandlers(
	r types.VersionedRouter,
	service Service,
	conf *config.Config,
	ncVersionSvc backend.NerdctlSystemSvc,
	logger flog.Logger,
) {
	h := newHandler(service, conf, ncVersionSvc, logger)
	r.HandleFunc("/info", h.info, http.MethodGet)
	r.HandleFunc("/version", h.version, http.MethodGet)
	r.HandleFunc("/_ping", h.ping, http.MethodHead, http.MethodGet)
	r.HandleFunc("/auth", h.auth, http.MethodPost)
	r.HandleFunc("/events", h.events, http.MethodGet)
}

func newHandler(service Service, conf *config.Config, ncSystemSvc backend.NerdctlSystemSvc, logger flog.Logger) *handler {
	return &handler{
		service:     service,
		Config:      conf,
		ncSystemSvc: ncSystemSvc,
		logger:      logger,
	}
}

type handler struct {
	service     Service
	Config      *config.Config
	ncSystemSvc backend.NerdctlSystemSvc
	logger      flog.Logger
}
