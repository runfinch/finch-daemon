// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/gorilla/mux"
	"github.com/moby/moby/api/server/httputils"
	"github.com/moby/moby/api/types/versions"
	"github.com/moby/moby/pkg/stdcopy"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) start(w http.ResponseWriter, r *http.Request) {
	var (
		execId               = mux.Vars(r)["id"]
		stdin                io.ReadCloser
		stdout, stderr       io.Writer
		printSuccessResponse func()
		attachHeaderWritten  = false
	)
	ctx := namespaces.WithNamespace(r.Context(), h.config.Namespace)

	cid, procId, err := parseExecId(execId)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewError(err))
		return
	}

	if r.Body == nil {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg("body should not be empty"))
		return
	}

	execStartCheck := &types.ExecStartCheck{}
	if err := json.NewDecoder(r.Body).Decode(execStartCheck); err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewError(fmt.Errorf("unable to parse request body: %w", err)))
		return
	}

	if !execStartCheck.Detach {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			response.JSON(w, http.StatusInternalServerError, response.NewErrorFromMsg("not a http.Hijacker"))
			return
		}

		conn, _, err := hijacker.Hijack()
		if err != nil {
			response.JSON(w, http.StatusInternalServerError, response.NewError(err))
			return
		}
		defer func() {
			if conn != nil {
				httputils.CloseStreams(conn)
			}
		}()

		// sets the connection to raw TCP mode. this will allow us to stream arbitrary data from the exec output over the connection
		_, err = conn.Write([]byte{})
		if err != nil {
			response.JSON(w, http.StatusInternalServerError, response.NewError(err))
			return
		}

		_, upgrade := r.Header["Upgrade"]
		successResponse := checkUpgradeStatus(ctx, upgrade)

		printSuccessResponse = func() {
			fmt.Fprint(conn, successResponse)
			// copy headers that were removed as part of hijack
			w.Header().WriteSubset(conn, nil)
			fmt.Fprint(conn, "\r\n")
			attachHeaderWritten = true
		}

		stdin = conn
		stdout = stdcopy.NewStdWriter(conn, stdcopy.Stdout)
		stderr = stdcopy.NewStdWriter(conn, stdcopy.Stderr)
	}

	startOptions := &types.ExecStartOptions{
		ExecStartCheck:  execStartCheck,
		ConID:           cid,
		ExecID:          procId,
		Stdin:           stdin,
		Stdout:          stdout,
		Stderr:          stderr,
		SuccessResponse: printSuccessResponse,
	}

	if err := h.service.Start(ctx, startOptions); err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		case errdefs.IsConflict(err):
			code = http.StatusConflict
		default:
			code = http.StatusInternalServerError
		}
		if execStartCheck.Detach {
			response.JSON(w, code, response.NewError(err))
			return
		}
		if !attachHeaderWritten {
			errResponse, _ := json.Marshal(response.NewError(err))
			fmt.Fprintf(stdout, "HTTP/1.1 %d %s\r\nContent-Type: application/json\r\n\r\n%s\r\n", code, http.StatusText(code), errResponse)
			return
		}
		stdout.Write([]byte(err.Error() + "\r\n"))
		return
	}

	if execStartCheck.Detach {
		response.Status(w, http.StatusOK)
	}
}

func checkUpgradeStatus(ctx context.Context, upgrade bool) string {
	contentType := "application/vnd.docker.raw-stream"
	if upgrade {
		if versions.GreaterThanOrEqualTo(httputils.VersionFromContext(ctx), "1.42") {
			contentType = "application/vnd.docker.multiplexed-stream"
		}
		return fmt.Sprintf("HTTP/1.1 101 UPGRADED\r\nContent-Type: %s\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n", contentType)
	} else {
		return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\n", contentType)
	}
}
