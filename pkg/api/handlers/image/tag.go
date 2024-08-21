// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/runfinch/finch-daemon/pkg/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) tag(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]
	repo := r.URL.Query().Get("repo")
	tag := r.URL.Query().Get("tag")
	err := h.service.Tag(r.Context(), name, repo, tag)
	if errdefs.IsNotFound(err) {
		response.JSON(w, http.StatusNotFound, response.NewError(err))
		return
	} else if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
		return
	}
	response.JSON(w, http.StatusCreated, "No error")
}
