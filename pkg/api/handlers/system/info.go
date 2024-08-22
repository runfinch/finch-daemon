// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"net/http"

	"github.com/runfinch/finch-daemon/pkg/api/response"
)

func (h *handler) info(w http.ResponseWriter, r *http.Request) {
	infoCompat, err := h.service.GetInfo(r.Context(), h.Config)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
		return
	}

	response.JSON(w, http.StatusOK, infoCompat)
}
