// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// package volume defines the API service and handlers for the volumes API.
package volume

import (
	"context"
	"net/http"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

//go:generate mockgen --destination=../../../mocks/mocks_volume/volumesvc.go -package=mocks_volume github.com/runfinch/finch-daemon/api/handlers/volume Service
type Service interface {
	Create(ctx context.Context, name string, labels []string) (*native.Volume, error)
	List(ctx context.Context, filters []string) (*types.VolumesListResponse, error)
	Remove(ctx context.Context, volName string, force bool) error
	Inspect(volName string) (*native.Volume, error)
}

func RegisterHandlers(r types.VersionedRouter, service Service, conf *config.Config, logger flog.Logger) {
	h := newHandler(service, conf, logger)

	r.SetPrefix("/volumes")
	r.HandleFunc("", h.list, http.MethodGet)
	r.HandleFunc("/{name:.*}", h.inspect, http.MethodGet)
	r.HandleFunc("/{name:.*}", h.remove, http.MethodDelete)
	r.HandleFunc("/create", h.create, http.MethodPost)
}

func newHandler(service Service, conf *config.Config, logger flog.Logger) *handler {
	return &handler{
		service: service,
		Config:  conf,
		logger:  logger,
	}
}

type handler struct {
	service Service
	Config  *config.Config
	logger  flog.Logger
}
