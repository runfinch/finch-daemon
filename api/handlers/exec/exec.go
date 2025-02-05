// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/containerd/nerdctl/v2/pkg/config"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

//go:generate mockgen --destination=../../../mocks/mocks_exec/execsvc.go -package=mocks_exec github.com/runfinch/finch-daemon/api/handlers/exec Service
type Service interface {
	Start(ctx context.Context, options *types.ExecStartOptions) error
	Resize(ctx context.Context, options *types.ExecResizeOptions) error
	Inspect(ctx context.Context, conId string, execId string) (*types.ExecInspect, error)
}

// RegisterHandlers registers all the supported endpoints related to exec APIs.
func RegisterHandlers(r types.VersionedRouter, service Service, conf *config.Config, logger flog.Logger) {
	h := newHandler(service, conf, logger)

	r.SetPrefix("/exec")
	r.HandleFunc("/{id:.*}/start", h.start, http.MethodPost)
	r.HandleFunc("/{id:.*}/resize", h.resize, http.MethodPost)
	r.HandleFunc("/{id:.*}/json", h.inspect, http.MethodGet)
}

// newHandler creates the handler that serves all exec APIs.
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

// parseExecId breaks down execId into a container ID and process ID.
func parseExecId(execId string) (string, string, error) {
	splitId := strings.Split(execId, "/")
	if len(splitId) != 2 {
		return "", "", fmt.Errorf("invalid exec id: %s", execId)
	}

	return splitId[0], splitId[1], nil
}
