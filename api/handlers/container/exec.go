// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/gorilla/mux"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// exec creates a new exec instance.
func (h *handler) exec(w http.ResponseWriter, r *http.Request) {
	cid, ok := mux.Vars(r)["id"]
	if !ok {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg("must specify a container id"))
		return
	}
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)

	if r.Body == nil {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg("request body should not be empty"))
		return
	}

	config := &types.ExecConfig{}
	if err := json.NewDecoder(r.Body).Decode(config); err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewError(fmt.Errorf("unable to parse request body: %w", err)))
		return
	}

	eid, err := h.service.ExecCreate(ctx, cid, *config)
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		case errdefs.IsConflict(err):
			code = http.StatusConflict
		default:
			code = http.StatusInternalServerError
		}
		response.JSON(w, code, response.NewError(err))
		return
	}

	response.JSON(w, http.StatusCreated, &types.ExecCreateResponse{
		Id: eid,
	})
}
