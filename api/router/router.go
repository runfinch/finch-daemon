// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/containerd/nerdctl/v2/pkg/config"
	ghandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/moby/moby/api/server/httputils"
	"github.com/moby/moby/api/types/versions"

	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/runfinch/finch-daemon/api/handlers/builder"
	"github.com/runfinch/finch-daemon/api/handlers/container"
	"github.com/runfinch/finch-daemon/api/handlers/distribution"
	"github.com/runfinch/finch-daemon/api/handlers/exec"
	"github.com/runfinch/finch-daemon/api/handlers/image"
	"github.com/runfinch/finch-daemon/api/handlers/network"
	"github.com/runfinch/finch-daemon/api/handlers/system"
	"github.com/runfinch/finch-daemon/api/handlers/volume"
	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/pkg/flog"
	"github.com/runfinch/finch-daemon/version"
)

var errRego = errors.New("error in rego policy file")
var errInput = errors.New("error in HTTP request")

type inputRegoRequest struct {
	Method string
	Path   string
}

// Options defines the router options to be passed into the handlers.
type Options struct {
	Config              *config.Config
	ContainerService    container.Service
	ImageService        image.Service
	NetworkService      network.Service
	SystemService       system.Service
	BuilderService      builder.Service
	VolumeService       volume.Service
	ExecService         exec.Service
	DistributionService distribution.Service
	RegoFilePath        string

	// NerdctlWrapper wraps the interactions with nerdctl to build
	NerdctlWrapper *backend.NerdctlWrapper
}

// New creates a new router and registers the handlers to it. Returns a handler object
// The struct definitions of the HTTP responses come from https://github.com/moby/moby/tree/master/api/types.
func New(opts *Options) (http.Handler, error) {
	r := mux.NewRouter()
	r.Use(VersionMiddleware)

	logger := flog.NewLogrus()

	if opts.RegoFilePath != "" {
		regoMiddleware, err := CreateRegoMiddleware(opts.RegoFilePath, logger)
		if err != nil {
			return nil, err
		}
		r.Use(regoMiddleware)
	}
	vr := types.VersionedRouter{Router: r}
	system.RegisterHandlers(vr, opts.SystemService, opts.Config, opts.NerdctlWrapper, logger)
	image.RegisterHandlers(vr, opts.ImageService, opts.Config, logger)
	container.RegisterHandlers(vr, opts.ContainerService, opts.Config, logger)
	network.RegisterHandlers(vr, opts.NetworkService, opts.Config, logger)
	builder.RegisterHandlers(vr, opts.BuilderService, opts.Config, logger, opts.NerdctlWrapper)
	volume.RegisterHandlers(vr, opts.VolumeService, opts.Config, logger)
	exec.RegisterHandlers(vr, opts.ExecService, opts.Config, logger)
	distribution.RegisterHandlers(vr, opts.DistributionService, opts.Config, logger)
	return ghandlers.LoggingHandler(os.Stderr, r), nil
}

// VersionMiddleware checks for the requested version of the api and makes sure it falls within the bounds
// of the supported version.
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

// CreateRegoMiddleware dynamically parses the rego file at the path specified in options
// and return a function that allows or denies the request based on the policy.
// Will return a nil function and an error if the given file path is blank or invalid.
func CreateRegoMiddleware(regoFilePath string, logger *flog.Logrus) (func(next http.Handler) http.Handler, error) {
	if regoFilePath == "" {
		return nil, errRego
	}

	query := "data.finch.authz.allow"
	nr := rego.New(
		rego.Load([]string{regoFilePath}, nil),
		rego.Query(query),
	)

	preppedQuery, err := nr.PrepareForEval(context.Background())
	if err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			input := inputRegoRequest{
				Method: r.Method,
				Path:   r.URL.Path,
			}

			logger.Debugf("OPA input being evaluated: Method=%s, Path=%s", input.Method, input.Path)

			rs, err := preppedQuery.Eval(r.Context(), rego.EvalInput(input))
			if err != nil {
				logger.Errorf("OPA policy evaluation failed: %v", err)
				response.SendErrorResponse(w, http.StatusInternalServerError, errInput)
				return
			}

			logger.Debugf("OPA evaluation results: %+v", rs)

			if !rs.Allowed() {
				logger.Infof("OPA request denied: Method=%s, Path=%s", r.Method, r.URL.Path)
				response.SendErrorResponse(w, http.StatusForbidden,
					fmt.Errorf("method %s not allowed for path %s", r.Method, r.URL.Path))
				return
			}
			logger.Debugf("OPA request allowed: Method=%s, Path=%s", r.Method, r.URL.Path)
			newReq := r.WithContext(r.Context())
			next.ServeHTTP(w, newReq)
		})
	}, nil
}
