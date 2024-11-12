// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/gorilla/mux"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) stats(w http.ResponseWriter, r *http.Request) {
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)

	stream, err := strconv.ParseBool(r.URL.Query().Get("stream"))
	if err != nil {
		stream = true // stream is true by default
	}

	cid := mux.Vars(r)["id"]
	statsCh, err := h.service.Stats(ctx, cid)
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		default:
			code = http.StatusInternalServerError
		}
		h.logger.Debugf("Stats container API responding with error code. Status code %d, Message: %s", code, err)
		response.SendErrorResponse(w, code, err)
		return
	}

	// set http header and initialize json encoder and response writer
	f, ok := w.(http.Flusher)
	if !ok {
		response.SendErrorResponse(
			w,
			http.StatusInternalServerError,
			fmt.Errorf("http ResponseWriter is not a http Flusher"),
		)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	e := json.NewEncoder(w)

	// get the first set of stats object
	statsJSON, ok := <-statsCh
	if !ok {
		h.logger.Errorf("stats channel closed unexpectedly")
		return
	}
	err = e.Encode(*statsJSON)
	if err != nil {
		h.logger.Errorf("error encoding stats to json: %s", err)
		return
	}
	f.Flush()

	// if streaming is disabled, simply return
	if !stream {
		return
	}

	// continuously send stats updates as JSON objects
	for statsJSON := range statsCh {
		err = e.Encode(*statsJSON)
		if err != nil {
			h.logger.Errorf("error encoding stats to json: %s", err)
			return
		}
		f.Flush()
	}
}
