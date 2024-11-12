// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/gorilla/mux"

	"github.com/runfinch/finch-daemon/api/auth"
	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) push(w http.ResponseWriter, r *http.Request) {
	authCfg, err := auth.DecodeAuthConfig(r.Header.Get(auth.AuthHeader))
	if err != nil {
		response.SendErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to decode the auth header: %w", err))
		return
	}

	// start the push job and send status updates to the response writer as JSON stream
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)
	streamWriter := response.NewStreamWriter(w)
	result, err := h.service.Push(ctx, mux.Vars(r)["name"], r.URL.Query().Get("tag"), authCfg, streamWriter)
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		case errdefs.IsInvalidFormat(err):
			code = http.StatusBadRequest
		default:
			code = http.StatusInternalServerError
		}
		h.logger.Debugf("Push Image API failed. Status code %d, Message: %s", code, err)
		streamWriter.WriteError(code, err)
		return
	}

	// send push result as out-of-band aux data
	if result != nil {
		auxData, err := json.Marshal(result)
		if err != nil {
			return
		}
		streamWriter.WriteAux(auxData)
	}
}
