// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"net/http"
	"strconv"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/runfinch/finch-daemon/api/response"
)

func (h *handler) load(w http.ResponseWriter, r *http.Request) {
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)
	quiet, err := strconv.ParseBool(r.URL.Query().Get("quiet"))
	if err != nil {
		quiet = false
	}
	out := response.NewStreamWriter(w)
	err = h.service.Load(ctx, r.Body, out, quiet)
	if err != nil {
		out.WriteError(http.StatusInternalServerError, err)
		return
	}
}
