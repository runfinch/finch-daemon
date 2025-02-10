// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"encoding/json"
	"net/http"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/gorilla/mux"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// From https://github.com/moby/moby/blob/v23.0.3/api/types/types.go#L634-L638
// NetworkConnect represents the data to be used to connect a container to the network.
type networkConnect struct {
	Container string
	// TODO: EndpointConfig *network.EndpointSettings `json:",omitempty"`
}

func (h *handler) connect(w http.ResponseWriter, r *http.Request) {
	var req networkConnect
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewError(err))
		return
	}

	ctx := namespaces.WithNamespace(r.Context(), h.config.Namespace)
	err := h.service.Connect(ctx, mux.Vars(r)["id"], req.Container)
	if err != nil {
		if errdefs.IsNotFound(err) {
			response.JSON(w, http.StatusNotFound, response.NewError(err))
			return
		}
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
		return
	}
	response.Status(w, http.StatusOK)
}
