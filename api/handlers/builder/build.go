// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/nerdctl/v2/pkg/api/types"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/utility/maputility"
)

// build function is the http handler function for /build API.
func (h *handler) build(w http.ResponseWriter, r *http.Request) {
	streamWriter := response.NewStreamWriter(w)

	// create the build options based on passed parameter
	buildOptions, err := h.getBuildOptions(w, r, streamWriter)
	if err != nil {
		streamWriter.WriteError(http.StatusInternalServerError, err)
		return
	}

	// call the service to build
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)
	result, err := h.service.Build(ctx, buildOptions, r.Body)
	if err != nil {
		streamWriter.WriteError(http.StatusInternalServerError, err)
		return
	}

	// send build result as out-of-band aux data
	for _, buildImage := range result {
		auxData, err := json.Marshal(buildImage)
		if err != nil {
			return
		}
		streamWriter.WriteAux(auxData)
	}
}

// getBuildOptions creates the build option parameter from http request which is requires by nerdctl build function.
func (h *handler) getBuildOptions(w http.ResponseWriter, r *http.Request, stream io.Writer) (*types.BuilderBuildOptions, error) {
	bkHost, err := h.ncBuildSvc.GetBuildkitHost()
	if err != nil {
		h.logger.Warnf("Failed to get buildkit host: %v", err.Error())
		return nil, err
	}
	options := types.BuilderBuildOptions{
		// TODO: investigate - interestingly nerdctl prints all the build log in stderr for some reason.
		Stdout: stream,
		Stderr: stream,
		GOptions: types.GlobalCommandOptions{
			Debug:         h.Config.Debug,
			Address:       h.Config.Address,
			Namespace:     h.Config.Namespace,
			Snapshotter:   h.Config.Snapshotter,
			DataRoot:      h.Config.DataRoot,
			CgroupManager: h.Config.CgroupManager,
			HostsDir:      h.Config.HostsDir,
		},
		BuildKitHost: bkHost,
		Tag:          getQueryParamList(r, "t", nil),
		File:         getQueryParamStr(r, "dockerfile", "Dockerfile"),
		Target:       getQueryParamStr(r, "target", ""),
		Platform:     getQueryParamList(r, "platform", []string{}),
		Rm:           getQueryParamBool(r, "rm", true),
		Progress:     "auto",
	}

	argsQuery := r.URL.Query().Get("buildargs")
	if argsQuery != "" {
		buildargs := make(map[string]string)
		err := json.Unmarshal([]byte(argsQuery), &buildargs)
		if err != nil {
			return nil, fmt.Errorf("unable to parse buildargs query: %s", err)
		}
		options.BuildArgs = maputility.Flatten(buildargs, maputility.KeyEqualsValueFormat)
	}
	return &options, nil
}

// getQueryParamStr fetch string query parameter and returns default value if empty.
func getQueryParamStr(r *http.Request, paramName string, defaultValue string) string {
	val := r.URL.Query().Get(paramName)
	if val == "" {
		return defaultValue
	}
	return val
}

// getQueryParamBool fetch boolean query parameter and returns default value if empty.
func getQueryParamBool(r *http.Request, paramName string, defaultValue bool) bool {
	val := r.URL.Query().Get(paramName)
	if val == "" {
		return defaultValue
	}
	if boolValue, err := strconv.ParseBool(val); err != nil {
		return defaultValue
	} else {
		return boolValue
	}
}

// getQueryParamList fetch list of string query parameter and returns default value if empty.
func getQueryParamList(r *http.Request, paramName string, defaultValue []string) []string {
	params := r.URL.Query()
	if params == nil || params[paramName] == nil {
		return defaultValue
	}
	return params[paramName]
}
