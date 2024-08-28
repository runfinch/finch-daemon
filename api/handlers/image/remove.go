// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

const (
	// TODO: Figure out how to add removeResponseUntaggedImage to the response.
	removeResponseUntaggedKey = "Untagged"
	removeResponseDeletedKey  = "Deleted"
)

func (h *handler) remove(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	f := r.URL.Query().Get("force")
	force, err := strconv.ParseBool(f)
	if err != nil {
		force = false
	}
	deleted, untagged, err := h.service.Remove(r.Context(), name, force)
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
		response.SendErrorResponse(w, code, err)
		return
	}
	response.JSON(w, http.StatusOK, h.buildRemoveResp(untagged, deleted))
}

func (handler) buildRemoveResp(untagged, deleted []string) []map[string]string {
	resp := make([]map[string]string, 0, len(deleted)+len(untagged))
	push := func(key string, items []string) {
		for _, item := range items {
			resp = append(resp, map[string]string{key: item})
		}
	}
	push(removeResponseUntaggedKey, untagged)
	push(removeResponseDeletedKey, deleted)
	return resp
}
