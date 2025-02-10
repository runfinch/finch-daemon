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

func (h *handler) remove(w http.ResponseWriter, r *http.Request) {
	cid := mux.Vars(r)["id"]
	volumeFlag, err := strconv.ParseBool(r.URL.Query().Get("v"))
	if err != nil {
		volumeFlag = false
	}
	forceFlag, err := strconv.ParseBool(r.URL.Query().Get("force"))
	if err != nil {
		forceFlag = false
	}
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)
	err = h.service.Remove(ctx, cid, forceFlag, volumeFlag)
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
		logrus.Debugf("Remove container API responding with error code. Status code %d, Message: %s", code, err.Error())
		response.SendErrorResponse(w, code, err)

		return
	}
	// successfully deleted. Send no content status.
	response.Status(w, http.StatusNoContent)
}
