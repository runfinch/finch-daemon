// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"net/http"

	"github.com/runfinch/finch-daemon/pkg/api/response"
)

// version is the basic form of GET `/version`, this allows docker.from_env() to work
// and allows testing with the docker python SDK directly
//
// TODO: Add in additional server information.
func (h *handler) version(w http.ResponseWriter, r *http.Request) {
	vInfo, err := h.service.GetVersion(r.Context())
	if err != nil {
		h.logger.Warnf("unable to retrieve server component versions: %v", err)
		response.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	response.JSON(w, http.StatusOK, vInfo)
}
