// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/runfinch/finch-daemon/api/response"
)

type CreateVolumeRequest struct {
	Name       string
	Driver     string
	DriverOpts map[string]string
	Labels     map[string]string
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	// https://docs.docker.com/engine/api/v1.41/#tag/Volume/operation/VolumeCreate
	var requestJson CreateVolumeRequest
	err := json.NewDecoder(r.Body).Decode(&requestJson)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
		return
	}

	if requestJson.Driver != "local" && requestJson.Driver != "" {
		fmt.Printf("driver is = %s\n", requestJson.Driver)
		h.logger.Warnf("Driver is not currently supported, ignoring")
	}

	if len(requestJson.DriverOpts) != 0 {
		h.logger.Warnf("Driver Options is not currently supported, ignoring\n")
	}

	labelSlice := labelMapToSlice(requestJson.Labels)
	vol, err := h.service.Create(r.Context(), requestJson.Name, labelSlice)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
		return
	}
	response.JSON(w, http.StatusOK, vol)
}

func labelMapToSlice(inputMap map[string]string) []string {
	labelSlice := []string{}
	for key, val := range inputMap {
		labelSlice = append(labelSlice, key+"="+val)
	}
	return labelSlice
}
