// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// inspect handles the api call and parses the one "id" variable.
func (h *handler) inspect(w http.ResponseWriter, r *http.Request) {
	nid := mux.Vars(r)["id"]
	if nid == "" {
		response.JSON(w, http.StatusInternalServerError, response.NewErrorFromMsg("id cannot be empty"))
		return
	}
	resp, err := h.service.Inspect(r.Context(), nid)
	if err != nil {
		if errdefs.IsNotFound(err) {
			response.JSON(w, http.StatusNotFound, response.NewError(err))
			return
		}
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
		return
	}
	response.JSON(w, http.StatusOK, resp)
}
