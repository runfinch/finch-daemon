// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"net/http"
	"strconv"
	"time"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/gorilla/mux"
	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) restart(w http.ResponseWriter, r *http.Request) {
	cid := mux.Vars(r)["id"]
	t, err := strconv.ParseInt(r.URL.Query().Get("t"), 10, 64)
	if err != nil {
		t = 10 // Docker/nerdctl default
	}
	timeout := time.Second * time.Duration(t)

	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)
	err = h.service.Restart(ctx, cid, timeout)
	// map the error into http status code and send response.
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		default:
			code = http.StatusInternalServerError
		}
		h.logger.Debugf("Restart container API responding with error code. Status code %d, Message: %s", code, err.Error())
		response.SendErrorResponse(w, code, err)
		return
	}
	// successfully restarted the container. Send no content status.
	response.Status(w, http.StatusNoContent)
}
