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
		psArgs = "-ef"
	}

	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)

	var buf bytes.Buffer

	globalOpt := ncTypes.GlobalCommandOptions(*h.Config)
	options := ncTypes.ContainerTopOptions{
		GOptions: globalOpt,
		Stdout:   &buf,
		PsArgs:   psArgs,
	}

	h.logger.Infof("calling nerdctl top with the following option : %s", options.PsArgs)
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
	titles, processes, err := parseTopOutput(lines)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.NewErrorFromMsg("invalid top output format"))
		return
	}

	resp := container.ContainerTopOKBody{
		Processes: processes,
		Titles:    titles,
	}

	response.JSON(w, http.StatusOK, resp)
}

func parseTopOutput(lines []string) ([]string, [][]string, error) {
	if len(lines) < 2 {
		return nil, nil, fmt.Errorf("insufficient output lines")
	}

	// Parse titles from the first line
	titles := strings.Fields(lines[0])
	if len(titles) == 0 {
		return nil, nil, fmt.Errorf("no titles found")
	}

	commandIndex := -1
	for i, name := range titles {
		if name == "CMD" || name == "COMMAND" || name == "ARGS" {
			commandIndex = i
			break
		}
	}

	// Parse processes
	processes := make([][]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		processMap := make(map[string]string)

		// Handle command field specially - join remaining fields
		if commandIndex != -1 && len(fields) > commandIndex {
			// Map fields before command field
			for i := 0; i < commandIndex && i < len(fields); i++ {
				if i < len(titles) {
					processMap[titles[i]] = fields[i]
				}
			}

			// Join command field and all remaining fields
			if commandIndex < len(titles) {
				commandValue := strings.Join(fields[commandIndex:], " ")
				processMap[titles[commandIndex]] = commandValue
			}
		} else {
			// No command field, map fields directly
			for i, field := range fields {
				if i < len(titles) {
					processMap[titles[i]] = field
				}
			}
		}

		// Build process array in the same order as titles
		process := make([]string, len(titles))
		for i, title := range titles {
			if value, exists := processMap[title]; exists {
				process[i] = value
			} else {
				process[i] = ""
			}
		}

		processes = append(processes, process)
	}

	return titles, processes, nil
}
