// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/moby/moby/api/server/httputils"
	"github.com/runfinch/finch-daemon/pkg/api/response"
	"github.com/runfinch/finch-daemon/pkg/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) putArchive(w http.ResponseWriter, r *http.Request) {
	cid := mux.Vars(r)["id"]
	path := r.URL.Query().Get("path")
	if path == "" {
		h.logger.Error("error handling request, bad path")
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg("must specify a file or directory path"))
		return
	}

	opts := &types.PutArchiveOptions{
		ContainerId: cid,
		Path:        path,
		Overwrite:   httputils.BoolValue(r, "noOverwriteDirNonDir"),
		CopyUIDGID:  httputils.BoolValue(r, "copyUIDGID"),
	}
	err := h.service.ExtractArchiveInContainer(r.Context(), opts, r.Body)
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		case errdefs.IsForbiddenError(err):
			code = http.StatusForbidden
		case errdefs.IsInvalidFormat(err):
			code = http.StatusBadRequest
		default:
			code = http.StatusInternalServerError
		}
		h.logger.Errorf("error handling request %v", err)
		response.SendErrorResponse(w, code, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
