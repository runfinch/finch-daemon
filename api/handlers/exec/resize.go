// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/gorilla/mux"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) resize(w http.ResponseWriter, r *http.Request) {
	execId := mux.Vars(r)["id"]
	ctx := namespaces.WithNamespace(r.Context(), h.config.Namespace)
	height, err := getQueryParamInt(r, "h")
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewError(err))
		return
	}
	width, err := getQueryParamInt(r, "w")
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewError(err))
		return
	}

	cid, procId, err := parseExecId(execId)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewError(err))
		return
	}

	resizeOptions := &types.ExecResizeOptions{
		ConID:  cid,
		ExecID: procId,
		Height: height,
		Width:  width,
	}
	if err := h.service.Resize(ctx, resizeOptions); err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		default:
			code = http.StatusInternalServerError
		}
		response.JSON(w, code, response.NewError(err))
		return
	}

	response.Status(w, http.StatusOK)
}

// getQueryParamInt fetches an integer query parameter and throws an error if empty.
func getQueryParamInt(r *http.Request, paramName string) (int, error) {
	val := r.URL.Query().Get(paramName)
	if val == "" {
		return 0, fmt.Errorf("query parameter %s required", paramName)
	}
	if intValue, err := strconv.Atoi(val); err != nil {
		return 0, fmt.Errorf("%s must be an integer", paramName)
	} else {
		return intValue, nil
	}
}
