// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package distribution

import (
	"context"
	"fmt"
	"net/http"

	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/nerdctl/pkg/config"
	dockertypes "github.com/docker/cli/cli/config/types"
	registrytypes "github.com/docker/docker/api/types/registry"
	"github.com/gorilla/mux"
	"github.com/runfinch/finch-daemon/api/auth"
	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

//go:generate mockgen --destination=../../../mocks/mocks_distribution/distributionsvc.go -package=mocks_distribution github.com/runfinch/finch-daemon/api/handlers/distribution Service
type Service interface {
	Inspect(ctx context.Context, name string, authCfg *dockertypes.AuthConfig) (*registrytypes.DistributionInspect, error)
}

func RegisterHandlers(r types.VersionedRouter, service Service, conf *config.Config, logger flog.Logger) {
	h := newHandler(service, conf, logger)
	r.HandleFunc("/distribution/{name}/json", h.inspect, http.MethodGet)
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

func (h *handler) inspect(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	// get auth creds from header
	authCfg, err := auth.DecodeAuthConfig(r.Header.Get(auth.AuthHeader))
	if err != nil {
		response.SendErrorResponse(w, http.StatusBadRequest, fmt.Errorf("failed to decode auth header: %s", err))
		return
	}
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)
	inspectRes, err := h.service.Inspect(ctx, name, authCfg)
	// map the error into http status code and send response.
	if err != nil {
		var code int
		switch {
		case errdefs.IsInvalidFormat(err):
			code = http.StatusBadRequest
		case errdefs.IsForbiddenError(err):
			code = http.StatusForbidden
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		default:
			code = http.StatusInternalServerError
		}
		h.logger.Debugf("Inspect Distribution API failed. Status code %d, Message: %s", code, err)
		response.SendErrorResponse(w, code, err)
		return
	}

	// return JSON response
	response.JSON(w, http.StatusOK, inspectRes)
}
