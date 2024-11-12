// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"net/http"

	"github.com/containerd/nerdctl/v2/pkg/config"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

//go:generate mockgen --destination=../../../mocks/mocks_network/networksvc.go -package=mocks_network github.com/runfinch/finch-daemon/api/handlers/network Service
type Service interface {
	Create(ctx context.Context, request types.NetworkCreateRequest) (types.NetworkCreateResponse, error)
	Connect(ctx context.Context, networkId, containerId string) error
	Inspect(ctx context.Context, networkId string) (*types.NetworkInspectResponse, error)
	Remove(ctx context.Context, networkId string) error
	List(ctx context.Context) ([]*types.NetworkInspectResponse, error)
}

// RegisterHandlers register all the supported endpoints related to the network APIs.
func RegisterHandlers(r types.VersionedRouter, service Service, conf *config.Config, logger flog.Logger) {
	h := newHandler(service, conf, logger)

	r.SetPrefix("/networks")
	r.HandleFunc("/create", h.create, http.MethodPost)
	r.HandleFunc("/{id:.*}/connect", h.connect, http.MethodPost)
	r.HandleFunc("/{id}", h.inspect, http.MethodGet)
	r.HandleFunc("/{id}", h.remove, http.MethodDelete)
	r.HandleFunc("/", h.list, http.MethodGet)
	r.HandleFunc("", h.list, http.MethodGet)
}

func newHandler(service Service, conf *config.Config, logger flog.Logger) *handler {
	return &handler{
		service: service,
		config:  conf,
		logger:  logger,
	}
}

type handler struct {
	service Service
	config  *config.Config
	logger  flog.Logger
}
