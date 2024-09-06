// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/moby/moby/api/server/httputils"
	"github.com/moby/moby/api/types/versions"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// attach handles the http request for attaching containers through hijacking the connection
// Modified from https://github.com/moby/moby/blob/5a9201ff477dd0f855d8f7fe59e9d59c4d90ac37/api/server/router/container/container_routes.go#L643.
//
// TODO: Add "Currently only one attach session is allowed." to the API doc.
func (h *handler) attach(w http.ResponseWriter, r *http.Request) {
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

	_, upgrade := r.Header["Upgrade"]
	contentType, successResponse := checkUpgradeStatus(r.Context(), upgrade)

	// define setupStreams to pass the connection, the stopchannel, and the success response
	setupStreams := func() (io.Writer, io.Writer, chan os.Signal, func(), error) {
		return conn, conn, stopChannel, func() {
			fmt.Fprint(conn, successResponse)
		}, nil
	}

	opts := &types.AttachOptions{
		GetStreams: setupStreams,
		UseStdin:   httputils.BoolValue(r, "stdin"),
		UseStdout:  httputils.BoolValue(r, "stdout"),
		UseStderr:  httputils.BoolValue(r, "stderr"),
		Logs:       httputils.BoolValue(r, "logs"),
		Stream:     httputils.BoolValue(r, "stream"),
		// TODO: implement DetachKeys now that David's nerdctl detachkeys is implemented
		// DetachKeys: r.URL.Query().Get("detachKeys"),
		// TODO: note that MuxStreams should be used in both in checkUpgradeStatus as well as
		// service.Attach, but since we always start containers in detached mode with tty=false,
		// whether the stream and the output will be multiplexed will always be true
		MuxStreams: true,
	}

	err = h.service.Attach(r.Context(), mux.Vars(r)["id"], opts)
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

// checkUpgradeStatus checks if the connection needs to be upgraded and returns the correct
// type and response.
func checkUpgradeStatus(ctx context.Context, upgrade bool) (string, string) {
	contentType := "application/vnd.docker.raw-stream"
	successResponse := fmt.Sprintf("HTTP/1.1 200 OK\r\n" +
		"Content-Type: application/vnd.docker.raw-stream\r\n\r\n")
	if upgrade {
		if versions.GreaterThanOrEqualTo(httputils.VersionFromContext(ctx), "1.42") {
			contentType = "application/vnd.docker.multiplexed-stream"
		}
		successResponse = fmt.Sprintf("HTTP/1.1 101 UPGRADED\r\nContent-Type: %s\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n", contentType)
	}
	return contentType, successResponse
}

// checkConnection monitors the hijacked connection and checks whether the connection is closed,
// running a closer function when it is closed.
//
// TODO: Refactor when we implement stdin.
func checkConnection(conn net.Conn, closer func()) {
	one := make([]byte, 1)
	if _, err := conn.Read(one); err == io.EOF {
		closer()
		conn.Close()
		return
	}
}
