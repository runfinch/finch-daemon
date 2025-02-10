// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"net/http"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/gorilla/mux"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) inspect(w http.ResponseWriter, r *http.Request) {
	execId := mux.Vars(r)["id"]
	ctx := namespaces.WithNamespace(r.Context(), h.config.Namespace)
	conId, procId, err := parseExecId(execId)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewError(err))
		return
	}

	inspect, err := h.service.Inspect(ctx, conId, procId)
	if err != nil {
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

	response.JSON(w, http.StatusOK, inspect)
}
