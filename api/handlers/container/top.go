// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/cmd/container"
	"github.com/gorilla/mux"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) top(w http.ResponseWriter, r *http.Request) {
	cid, ok := mux.Vars(r)["id"]
	if !ok || cid == "" {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg("must specify a container ID"))
		return
	}

	psArgs := r.URL.Query().Get("ps_args")
	if psArgs == "" {
		// Set default ps arguments if none provided
		psArgs = "-ef" // or whatever default you want to use
	}

	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)

	var buf bytes.Buffer

	globalOpt := ncTypes.GlobalCommandOptions(*h.Config)
	options := ncTypes.ContainerTopOptions{
		GOptions: globalOpt,
		Stdout:   &buf,
		PsArgs:   psArgs,
	}

	fmt.Printf("calling nerdctl top with the following option : %s", options.PsArgs)
	err := h.service.Top(ctx, cid, options)
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		case errdefs.IsConflict(err):
			code = http.StatusConflict
		case strings.Contains(err.Error(), "unknown") || strings.Contains(err.Error(), "invalid"):
			code = http.StatusBadRequest
		default:
			code = http.StatusInternalServerError
		}
		response.JSON(w, code, response.NewError(err))
		return
	}

	// Parse the output
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 2 {
		response.JSON(w, http.StatusInternalServerError, response.NewErrorFromMsg("invalid top output format"))
		return
	}

	// Parse titles
	titles := strings.Fields(lines[0])

	// Find the CMD/COMMAND column index
	cmdIndex := -1
	for i, name := range titles {
		if name == "CMD" || name == "COMMAND" || name == "ARGS" {
			cmdIndex = i
			break
		}
	}

	// Parse processes
	processes := make([][]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		if len(strings.TrimSpace(line)) > 0 {
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}

			if cmdIndex != -1 && len(fields) > cmdIndex {
				process := make([]string, cmdIndex)
				copy(process, fields[:cmdIndex])
				process = append(process, strings.Join(fields[cmdIndex:], " "))
				processes = append(processes, process)
			} else {
				processes = append(processes, fields)
			}
		}
	}

	resp := container.ContainerTopOKBody{
		Processes: processes,
		Titles:    titles,
	}

	response.JSON(w, http.StatusOK, resp)
}
