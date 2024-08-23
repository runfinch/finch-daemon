// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"net/http"

	"github.com/runfinch/finch-daemon/version"
)

// ping is a simple API endpoint to verify the server's accessibility.
func (h *handler) ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("API-Version", version.DefaultApiVersion)
}
