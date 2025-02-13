// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"net/http"
	"strconv"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) inspect(w http.ResponseWriter, r *http.Request) {
	cid := mux.Vars(r)["id"]
	sizeflag, err := strconv.ParseBool(r.URL.Query().Get("size"))
	if err != nil {
		sizeflag = false
	}
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)
	c, err := h.service.Inspect(ctx, cid, sizeflag)
	// map the error into http status code and send response.
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		default:
			code = http.StatusInternalServerError
		}
		logrus.Debugf("Inspect Container API failed. Status code %d, Message: %s", code, err)
		response.SendErrorResponse(w, code, err)

		return
	}

	// return JSON response
	response.JSON(w, http.StatusOK, c)
}
