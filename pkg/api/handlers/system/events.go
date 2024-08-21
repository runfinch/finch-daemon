// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"encoding/json"
	"fmt"
	"net/http"

	eventtype "github.com/runfinch/finch-daemon/pkg/api/events"
	"github.com/runfinch/finch-daemon/pkg/api/response"
)

func (h *handler) events(w http.ResponseWriter, r *http.Request) {
	filters := make(map[string][]string)
	filterQuery := r.URL.Query().Get("filters")
	if filterQuery != "" {
		err := json.Unmarshal([]byte(filterQuery), &filters)
		if err != nil {
			response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg(fmt.Sprintf("invalid filter: %s", err)))
			return
		}
	}

	eventCh, errCh := h.service.SubscribeEvents(r.Context(), filters)

	encoder := json.NewEncoder(w)

	flusher := w.(http.Flusher)
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		var e *eventtype.Event
		select {
		case e = <-eventCh:
		case err := <-errCh:
			// logging on debug level because an error is expected when the client stops listening for events
			h.logger.Debugf("received error, exiting: %s", err)
			return
		}

		if e != nil {
			err := encoder.Encode(e)
			if err != nil {
				h.logger.Errorf("could not encode event to JSON: %s", err)
				return
			}
			flusher.Flush()
		}
	}
}
