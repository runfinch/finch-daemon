// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"net/http"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) rename(w http.ResponseWriter, r *http.Request) {
	cid := mux.Vars(r)["id"]
	newName := r.URL.Query().Get("name")
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)
	globalOpt := ncTypes.GlobalCommandOptions(*h.Config)
	opts := ncTypes.ContainerRenameOptions{
		GOptions: globalOpt,
		Stdout:   nil,
	}
	err := h.service.Rename(ctx, cid, newName, opts)
	// map the error into http status code and send response.
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
		logrus.Debugf("Rename container API responding with error code. Status code %d, Message: %v", code, err)
		response.SendErrorResponse(w, code, err)
		return
	}
	// successfully stopped. Send no content status.
	response.Status(w, http.StatusNoContent)
}
