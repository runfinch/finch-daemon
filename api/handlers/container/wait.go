// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"net/http"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

type waitError struct {
	Message string
}

type waitResp struct {
	StatusCode int64
	Error      *waitError `json:",omitempty"`
}

func (h *handler) wait(w http.ResponseWriter, r *http.Request) {
	cid := mux.Vars(r)["id"]

	// TODO: condition is not used because nerdctl doesn't support it
	condition := r.URL.Query().Get("condition")
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)
	code, err := h.service.Wait(ctx, cid, condition)

	if code == -1 {
		var errorCode int
		switch {
		case errdefs.IsNotFound(err):
			errorCode = http.StatusNotFound
		case errdefs.IsInvalidFormat(err):
			errorCode = http.StatusBadRequest
		default:
			errorCode = http.StatusInternalServerError
		}
		logrus.Debugf("Wait container API responding with error code. Status code %d, Message: %s", errorCode, err.Error())
		response.SendErrorResponse(w, errorCode, err)

		return
	}

	waitResponse := waitResp{
		StatusCode: code,
	}
	// if there is no err then don't need to set the error msg. e.g. when container is stopped the wait should return
	// {"Error":null,"StatusCode":0}
	if err != nil {
		waitResponse.Error = &waitError{Message: err.Error()}
	}

	response.JSON(w, http.StatusOK, waitResponse)
}
