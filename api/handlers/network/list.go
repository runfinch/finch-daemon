// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"net/http"

	"github.com/runfinch/finch-daemon/api/response"
)

// list handles the api call to list and returns a json object.
func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.List(r.Context())
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
		return
	}
	response.JSON(w, http.StatusOK, resp)
}
