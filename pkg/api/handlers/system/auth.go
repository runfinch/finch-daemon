// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/runfinch/finch-daemon/pkg/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

type authReq struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	ServerAddress string `json:"serveraddress"`
	// email is deprecated:
	// https://github.com/moby/moby/blob/0200623ef7b7b166c675cb14502cbc0704d3dfd4/api/types/registry/authconfig.go#L21-L24
}

type authResp struct {
	Status        string `json:"Status"`
	IdentityToken string `json:"IdentityToken"`
}

func (h *handler) auth(w http.ResponseWriter, r *http.Request) {
	var req authReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewError(err))
		return
	}
	if req.Username == "" {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg("username is required"))
		return
	}
	if req.Password == "" {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg("password is required"))
		return
	}

	token, err := h.service.Auth(r.Context(), req.Username, req.Password, req.ServerAddress)
	if err != nil {
		code := http.StatusInternalServerError
		if errdefs.IsUnauthenticated(err) {
			code = http.StatusUnauthorized
		}
		response.JSON(w, code, response.NewError(fmt.Errorf("failed to authenticate: %w", err)))
		return
	}
	response.JSON(w, http.StatusOK, &authResp{
		Status:        "Login Succeeded",
		IdentityToken: token,
	})
}
