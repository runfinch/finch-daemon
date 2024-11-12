// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package container comprises functions and structures related to container APIs
package container

import (
	"context"
	"io"
	"net/http"
	"time"

	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/config"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

//go:generate mockgen --destination=../../../mocks/mocks_container/containersvc.go -package=mocks_container github.com/runfinch/finch-daemon/api/handlers/container Service
type Service interface {
	GetPathToFilesInContainer(ctx context.Context, cid string, path string) (string, func(), error)
	Remove(ctx context.Context, cid string, force, removeVolumes bool) error
	Wait(ctx context.Context, cid string, condition string) (code int64, err error)
	Start(ctx context.Context, cid string) error
	Stop(ctx context.Context, cid string, timeout *time.Duration) error
	Restart(ctx context.Context, cid string, timeout time.Duration) error
	Create(ctx context.Context, image string, cmd []string, createOpt ncTypes.ContainerCreateOptions, netOpt ncTypes.NetworkOptions) (string, error)
	Inspect(ctx context.Context, cid string) (*types.Container, error)
	WriteFilesAsTarArchive(filePath string, writer io.Writer, slashDot bool) error
	Attach(ctx context.Context, cid string, opts *types.AttachOptions) error
	List(ctx context.Context, listOpts ncTypes.ContainerListOptions) ([]types.ContainerListItem, error)
	Rename(ctx context.Context, cid string, newName string, opts ncTypes.ContainerRenameOptions) error
	Logs(ctx context.Context, cid string, opts *types.LogsOptions) error
	ExtractArchiveInContainer(ctx context.Context, putArchiveOpt *types.PutArchiveOptions, body io.ReadCloser) error
	Stats(ctx context.Context, cid string) (<-chan *types.StatsJSON, error)
	ExecCreate(ctx context.Context, cid string, config types.ExecConfig) (string, error)
}

// RegisterHandlers register all the supported endpoints related to the container APIs.
func RegisterHandlers(r types.VersionedRouter, service Service, conf *config.Config, logger flog.Logger) {
	h := newHandler(service, conf, logger)

	r.SetPrefix("/containers")
	r.HandleFunc("/{id:.*}", h.remove, http.MethodDelete)
	r.HandleFunc("/{id:.*}/start", h.start, http.MethodPost)
	r.HandleFunc("/{id:.*}/stop", h.stop, http.MethodPost)
	r.HandleFunc("/{id:.*}/restart", h.restart, http.MethodPost)
	r.HandleFunc("/{id:.*}/remove", h.remove, http.MethodPost)
	r.HandleFunc("/{id:.*}/wait", h.wait, http.MethodPost)
	r.HandleFunc("/create", h.create, http.MethodPost)
	r.HandleFunc("/{id:.*}/json", h.inspect, http.MethodGet)
	r.HandleFunc("/{id:.*}/archive", h.getArchive, http.MethodGet)
	r.HandleFunc("/{id:.*}/attach", h.attach, http.MethodPost)
	r.HandleFunc("/json", h.list, http.MethodGet)
	r.HandleFunc("/{id:.*}/rename", h.rename, http.MethodPost)
	r.HandleFunc("/{id:.*}/logs", h.logs, http.MethodGet)
	r.HandleFunc("/{id:.*}/archive", h.putArchive, http.MethodPut)
	r.HandleFunc("/{id:.*}/stats", h.stats, http.MethodGet)
	r.HandleFunc("/{id:.*}/exec", h.exec, http.MethodPost)
}

// newHandler creates the handler that serves all the container related APIs.
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
