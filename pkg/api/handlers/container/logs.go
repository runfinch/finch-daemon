// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/moby/moby/api/server/httputils"
	"github.com/runfinch/finch-daemon/pkg/api/response"
	"github.com/runfinch/finch-daemon/pkg/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// logs handles the http request for attaching to a container's logs
func (h *handler) logs(w http.ResponseWriter, r *http.Request) {
	// return early if neither stdout and stderr are set
	stdout, stderr := httputils.BoolValue(r, "stdout"), httputils.BoolValue(r, "stderr")
	if !(stdout || stderr) {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg(
			"you must choose at least one stream"))
		return
	}

	// setup hijacker && hijack connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		response.JSON(w, http.StatusBadRequest, response.
			NewErrorFromMsg("the response writer is not a http.Hijacker"))
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
		return
	}

	// set raw mode
	_, err = conn.Write([]byte{})
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
		return
	}

	// setup stop channel to communicate with logviewer,
	stopChannel := make(chan os.Signal, 1)
	signal.Notify(stopChannel, syscall.SIGTERM, syscall.SIGINT)
	go checkConnection(conn, func() {
		stopChannel <- os.Interrupt
	})

	contentType, successResponse := checkUpgradeStatus(r.Context(), false)

	// define setupStreams to pass the connection, the stopchannel, and the success response
	setupStreams := func() (io.Writer, io.Writer, chan os.Signal, func(), error) {
		return conn, conn, stopChannel, func() {
			fmt.Fprintf(conn, successResponse)
		}, nil
	}

	opts := &types.LogsOptions{
		GetStreams: setupStreams,
		Stdout:     stdout,
		Stderr:     stderr,
		Follow:     httputils.BoolValueOrDefault(r, "follow", false),
		Since:      httputils.Int64ValueOrZero(r, "since"),
		Until:      httputils.Int64ValueOrZero(r, "until"),
		Timestamps: httputils.BoolValueOrDefault(r, "timestamps", false),
		Tail:       r.Form.Get("tail"),
		MuxStreams: true,
	}

	err = h.service.Logs(r.Context(), mux.Vars(r)["id"], opts)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errdefs.IsNotFound(err) {
			statusCode = http.StatusNotFound
		}
		statusText := http.StatusText(statusCode)
		fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\n"+
			"Content-Type: %s\r\n\r\n%s\r\n", statusCode, statusText, contentType, err.Error())
	}
	if conn != nil {
		httputils.CloseStreams(conn)
	}
}
