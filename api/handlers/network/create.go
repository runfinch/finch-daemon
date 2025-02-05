// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/containerd/containerd/v2/pkg/namespaces"

	respond "github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	ctx := namespaces.WithNamespace(r.Context(), h.config.Namespace)

	var request types.NetworkCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Errorf("Failed decoding create network request: %v", err)
		respond.SendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	isMissingARequiredField := func(r types.NetworkCreateRequest) bool {
		// The Docker API specification only requires network name.
		return r.Name == ""
	}

	if isMissingARequiredField(request) {
		// The request is valid JSON, but missing a required field.
		h.logger.Warn("Create network request received missing required field.")

		// Docker Engine returns 500 Internal Server Error in such an instance.
		respond.SendErrorResponse(w, http.StatusInternalServerError, errors.New("missing required field"))
		return
	}

	h.logger.Debugf("Create network '%s'.", request.Name)

	response, err := h.service.Create(ctx, request)
	if err != nil {
		h.handleCreateError(w, request, err)
		return
	}

	h.logger.Debugf("Network '%s' created.", request.Name)
	respond.JSON(w, http.StatusCreated, &response)
}

func (h *handler) handleCreateError(w http.ResponseWriter, request types.NetworkCreateRequest, err error) {
	var code int

	if errdefs.IsNotFound(err) {
		h.logger.Errorf("Create network '%s' failed for CNI plugin '%s' not supported.", request.Name, request.Driver)
		code = http.StatusNotFound
	} else {
		h.logger.Errorf("Create network '%s' failed: %v.", request.Name, err)
		code = http.StatusInternalServerError
	}

	respond.SendErrorResponse(w, code, err)
}
