// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package main implements a credential helper for Finch daemon
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

const (
	// CredentialSocketPath is the path to the credential socket.
	CredentialSocketPath = "/run/finch/credential.sock"

	// ConnectionTimeout is the timeout for connecting to the credential socket.
	ConnectionTimeout = 5 * time.Second
)

var log flog.Logger

// FinchCredentialHelper implements the credentials.Helper interface.
type FinchCredentialHelper struct{}

// Add is not implemented for Finch credential helper.
func (h FinchCredentialHelper) Add(*credentials.Credentials) error {
	return fmt.Errorf("not implemented")
}

// Delete is not implemented for Finch credential helper.
func (h FinchCredentialHelper) Delete(serverURL string) error {
	return fmt.Errorf("not implemented")
}

// List is not implemented for Finch credential helper.
func (h FinchCredentialHelper) List() (map[string]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// Get retrieves credentials from the Finch daemon.
func (h FinchCredentialHelper) Get(serverURL string) (string, string, error) {
	buildID := os.Getenv("FINCH_BUILD_ID")
	if buildID == "" {
		return "", "", credentials.NewErrCredentialsNotFound()
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", CredentialSocketPath)
			},
		},
	}

	// Create request with JSON body.
	reqBody := struct {
		BuildID    string `json:"buildID"`
		ServerAddr string `json:"serverAddr"`
	}{BuildID: buildID, ServerAddr: serverURL}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/finch/credentials", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("error reading response: %v", err)
		return "", "", fmt.Errorf("error reading response: %v", err)
	}

	var errorResponse struct {
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(responseBytes, &errorResponse); err == nil && errorResponse.Error != "" {
		return "", "", fmt.Errorf("error from credential service: %s", errorResponse.Error)
	}

	var authConfig dockertypes.AuthConfig
	if err := json.Unmarshal(responseBytes, &authConfig); err != nil {
		return "", "", fmt.Errorf("error parsing response: %v", err)
	}

	return authConfig.Username, authConfig.Password, nil
}

func main() {
	credentials.Serve(FinchCredentialHelper{})
}
