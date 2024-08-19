// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package router

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/containerd/nerdctl/pkg/config"
	ghandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/moby/moby/api/server/httputils"
	"github.com/moby/moby/api/types/versions"
	"github.com/runfinch/finch-daemon/pkg/api/handlers/builder"
	"github.com/runfinch/finch-daemon/pkg/api/handlers/container"
	"github.com/runfinch/finch-daemon/pkg/api/handlers/exec"
	"github.com/runfinch/finch-daemon/pkg/api/handlers/image"
	"github.com/runfinch/finch-daemon/pkg/api/handlers/network"
	"github.com/runfinch/finch-daemon/pkg/api/handlers/system"
	"github.com/runfinch/finch-daemon/pkg/api/handlers/volume"
	"github.com/runfinch/finch-daemon/pkg/api/response"
	"github.com/runfinch/finch-daemon/pkg/api/types"
	"github.com/runfinch/finch-daemon/pkg/backend"
	"github.com/runfinch/finch-daemon/pkg/flog"
	"github.com/runfinch/finch-daemon/pkg/version"
)

// Options defines the router options to be passed into the handlers
type Options struct {
	Config           *config.Config
	ContainerService container.Service
	ImageService     image.Service
	NetworkService   network.Service
	SystemService    system.Service
	BuilderService   builder.Service
	VolumeService    volume.Service
	ExecService      exec.Service

	// NerdctlWrapper wraps the interactions with nerdctl to build
	NerdctlWrapper *backend.NerdctlWrapper
}

// New creates a new router and registers the handlers to it. Returns a handler object
// The struct definitions of the HTTP responses come from https://github.com/moby/moby/tree/master/api/types.
func New(opts *Options) http.Handler {
	r := mux.NewRouter()
	r.Use(VersionMiddleware)
	vr := types.VersionedRouter{Router: r}

	logger := flog.NewLogrus()
	system.RegisterHandlers(vr, opts.SystemService, opts.Config, opts.NerdctlWrapper, logger)
	image.RegisterHandlers(vr, opts.ImageService, opts.Config, logger)
	container.RegisterHandlers(vr, opts.ContainerService, opts.Config, logger)
	network.RegisterHandlers(vr, opts.NetworkService, opts.Config, logger)
	builder.RegisterHandlers(vr, opts.BuilderService, opts.Config, logger, opts.NerdctlWrapper)
	volume.RegisterHandlers(vr, opts.VolumeService, opts.Config, logger)
	exec.RegisterHandlers(vr, opts.ExecService, opts.Config, logger)
	return ghandlers.LoggingHandler(os.Stderr, r)
}

// VersionMiddleware checks for the requested version of the api and makes sure it falls within the bounds
// of the supported version
func VersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ctx := context.WithValue(r.Context(), httputils.APIVersionKey{}, version.DefaultApiVersion)
		APIVersion, ok := vars["version"]
		if ok {
			if versions.LessThan(APIVersion, version.MinimumApiVersion) {
				response.SendErrorResponse(w, http.StatusBadRequest,
					fmt.Errorf("your api version, v%s, is below the minimum supported version, v%s",
						APIVersion, version.MinimumApiVersion))
				return
			} else if versions.GreaterThan(APIVersion, version.DefaultApiVersion) {
				response.SendErrorResponse(w, http.StatusBadRequest,
					fmt.Errorf("your api version, v%s, is newer than the server's version, v%s",
						APIVersion, version.DefaultApiVersion))
				return
			}
			ctx = context.WithValue(r.Context(), httputils.APIVersionKey{}, APIVersion)
		}
		newReq := r.WithContext(ctx)
		next.ServeHTTP(w, newReq)
	})
}
