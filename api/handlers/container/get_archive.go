// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) getArchive(w http.ResponseWriter, r *http.Request) {
	cid := mux.Vars(r)["id"]
	path := r.URL.Query().Get("path")
	if path == "" {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg("must specify a file or directory path"))
		return
	}

	filePath, cleanup, err := h.service.GetPathToFilesInContainer(r.Context(), cid, path)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		default:
			code = http.StatusInternalServerError
		}
		h.logger.Debugf("Responding with error. Error code: %d, Message: %s", code, err.Error())
		response.SendErrorResponse(w, code, err)
		return
	}

	// "/." is a Docker thing that instructions the copy command to download contents of the folder only
	pathHasSlashDot := strings.HasSuffix(path, string(os.PathSeparator)+".")

	w.Header().Set("Content-Type", "application/x-tar")
	// TODO: to be compatible with the docker CLI, we will need to implement the X-Docker-Container-Path-Stat header here
	// see https://github.com/docker/go-docker/blob/4daae26030ad00e348edddff9767924ae57a3b82/container_copy.go#L90 for
	// where the CLI uses it
	w.WriteHeader(http.StatusOK)
	// path.Join() removes "/." from the end of a path, so filePath will never end in "/.". therefore, we need to propagate
	// a bool that tells us whether the original path had a "/."
	err = h.service.WriteFilesAsTarArchive(filePath, w, pathHasSlashDot)
	if err != nil {
		h.logger.Errorf("Could not send response: %s\n", err)
	}
}
