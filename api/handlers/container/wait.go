// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

type waitError struct {
	Message string
}

type waitResp struct {
	StatusCode int
	Error      *waitError `json:",omitempty"`
}

type codeCapturingWriter struct {
	code int
}

func (w *codeCapturingWriter) Write(p []byte) (n int, err error) {
	w.code, _ = strconv.Atoi(strings.TrimSpace(string(p)))
	return len(p), nil
}

func (h *handler) wait(w http.ResponseWriter, r *http.Request) {
	cid := mux.Vars(r)["id"]

	// TODO: condition is not used because nerdctl doesn't support it
	condition := r.URL.Query().Get("condition")
	if condition != "" {
		logrus.Debugf("Wait container API called with condition options - not supported.")
		response.SendErrorResponse(w, http.StatusBadRequest, fmt.Errorf("wait condition is not supported"))
		return
	}
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)

	// Create the custom writer
	codeWriter := &codeCapturingWriter{}

	globalOpt := ncTypes.GlobalCommandOptions(*h.Config)
	options := ncTypes.ContainerWaitOptions{
		GOptions: globalOpt,
		Stdout:   codeWriter,
	}

	err := h.service.Wait(ctx, cid, options)

	code := codeWriter.code

	if err != nil {
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

	// if there is no err then don't need to set the error msg. e.g. when container is stopped the wait should return
	// {"Error":null,"StatusCode":0}
	waitResponse := waitResp{
		StatusCode: code,
	}

	response.JSON(w, http.StatusOK, waitResponse)
}
