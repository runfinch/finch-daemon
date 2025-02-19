// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (h *handler) start(w http.ResponseWriter, r *http.Request) {
	cid := mux.Vars(r)["id"]
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)

	detachKeys := r.URL.Query().Get("detachKeys")

	if err := validateDetachKeys(detachKeys); err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg(fmt.Sprintf("Invalid detach keys: %v", err)))
		return
	}

	devNull, err := os.OpenFile("/dev/null", os.O_WRONLY, 0600)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg("failed to open /dev/null"))
		return
	}
	defer devNull.Close()

	globalOpt := ncTypes.GlobalCommandOptions(*h.Config)
	options := ncTypes.ContainerStartOptions{
		GOptions:   globalOpt,
		Stdout:     devNull,
		Attach:     false,
		DetachKeys: detachKeys,
	}

	err = h.service.Start(ctx, cid, options)
	// map the error into http status code and send response.
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		case errdefs.IsNotModified(err):
			code = http.StatusNotModified
		default:
			code = http.StatusInternalServerError
		}
		logrus.Debugf("Start container API responding with error code. Status code %d, Message: %s", code, err.Error())
		response.SendErrorResponse(w, code, err)
		return
	}
	// successfully started the container. Send no content status.
	response.Status(w, http.StatusNoContent)
}

func validateDetachKeys(detachKeys string) error {
	if detachKeys == "" {
		return nil // Empty string is valid (use default)
	}

	parts := strings.Split(detachKeys, ",")
	for _, part := range parts {
		if err := validateSingleKey(part); err != nil {
			return err
		}
	}

	return nil
}

func validateSingleKey(key string) error {
	key = strings.TrimSpace(key)
	if len(key) == 1 {
		// Single character key
		if !unicode.IsLetter(rune(key[0])) {
			return fmt.Errorf("invalid key: %s - single character must be a letter", key)
		}
	} else if strings.HasPrefix(key, "ctrl-") {
		ctrlKey := strings.TrimPrefix(key, "ctrl-")
		if len(ctrlKey) != 1 {
			return fmt.Errorf("invalid ctrl key: %s - must be a single character", ctrlKey)
		}
		validCtrlKeys := "abcdefghijklmnopqrstuvwxyz@[\\]^_"
		if !strings.Contains(validCtrlKeys, strings.ToLower(ctrlKey)) {
			return fmt.Errorf("invalid ctrl key: %s - must be one of %s", ctrlKey, validCtrlKeys)
		}
	} else {
		return fmt.Errorf("invalid key: %s - must be a single character or ctrl- combination", key)
	}
	return nil
}
