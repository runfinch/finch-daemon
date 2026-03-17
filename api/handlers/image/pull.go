// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/containerd/containerd/v2/pkg/namespaces"

	"github.com/runfinch/finch-daemon/api/auth"
	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// The /images/create API pulls the image specified by given name and tag.
// Importing images is not supported.
func (h *handler) pull(w http.ResponseWriter, r *http.Request) {
	warnings := handleUnsupportedParams(r)
	for warning := range warnings {
		h.logger.Warn(warning)
	}

	// get auth creds from header
	authCfg, err := auth.DecodeAuthConfig(r.Header.Get(auth.AuthHeader))
	if err != nil {
		response.SendErrorResponse(w, http.StatusBadRequest, fmt.Errorf("failed to decode auth header: %s", err))
		return
	}

	// image name
	name, tag, err := parseNameAndTag(r)
	if err != nil {
		response.SendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	platform := r.URL.Query().Get("platform")

	// start the pull job and send status updates to the response writer as JSON stream
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)
	streamWriter := response.NewPullJobWriter(w)
	err = h.service.Pull(ctx, name, tag, platform, authCfg, streamWriter)
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		case errdefs.IsInvalidFormat(err):
			code = http.StatusBadRequest
		default:
			code = http.StatusInternalServerError
		}
		h.logger.Debugf("Create Image API failed. Status code %d, Message: %s", code, err)
		streamWriter.WriteError(code, err)
		return
	}
	streamWriter.Write([]byte(fmt.Sprintf("Pulled %s:%s\n", name, tag)))
}

func handleUnsupportedParams(r *http.Request) []string {
	// unsupported query parameters: fromSrc, repo, message, changes.
	// fromSrc, repo, and message are only used when importing images.
	warnings := []string{}
	if r.URL.Query().Get("fromSrc") != "" {
		warnings = append(warnings, "fromSrc parameter specified but importing images is not supported")
	}
	if r.URL.Query().Get("repo") != "" {
		warnings = append(warnings, "repo parameter specified but importing images is not supported")
	}
	if r.URL.Query().Get("message") != "" {
		warnings = append(warnings, "message parameter specified but importing images is not supported")
	}
	if r.URL.Query().Get("changes") != "" {
		warnings = append(warnings, "changes parameter is not supported")
	}

	return warnings
}

var splitRE = regexp.MustCompile(`[@:]`)

func parseNameAndTag(r *http.Request) (string, string, error) {
	// image name
	nameParam := r.URL.Query().Get("fromImage")
	if nameParam == "" {
		return "", "", fmt.Errorf("fromImage must be specified")
	}

	tagParam := r.URL.Query().Get("tag")

	// If an explicit tag param is provided, use it directly.
	// Strip any "@digest" suffix from fromImage (e.g. Docker sends
	// "fromImage=alpine@sha256:abc&tag=sha256:abc" for digest pulls).
	// We only strip at '@', not ':', to preserve registry URLs like "localhost:5000/image".
	if tagParam != "" {
		name := nameParam
		if idx := strings.IndexByte(nameParam, '@'); idx >= 0 {
			name = nameParam[:idx]
		}
		if name == "" {
			return "", "", fmt.Errorf("invalid image: %s", nameParam)
		}
		return name, tagParam, nil
	}

	// No explicit tag param: parse the inline tag or digest from fromImage (split on '@' or ':').
	parts := splitRE.Split(nameParam, 2)
	name := parts[0]
	if name == "" {
		return "", "", fmt.Errorf("invalid image: %s", nameParam)
	}
	var tag string
	if len(parts) > 1 {
		tag = parts[1]
	}
	if tag == "" {
		return "", "", fmt.Errorf("image tag/digest must be specified")
	}

	return name, tag, nil
}
