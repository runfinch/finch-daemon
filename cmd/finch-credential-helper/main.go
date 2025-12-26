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
	"path/filepath"
	"strings"
	"time"

	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/runfinch/finch-daemon/pkg/config"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

const (
	// ConnectionTimeout is the timeout for connecting to the credential socket.
	ConnectionTimeout = 5 * time.Second
	BufferSize = 4096
)

// CredentialSocketPath is the path to the credential socket.
var CredentialSocketPath = config.DefaultCredentialAddr

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
	if os.Getenv("USE_NATIVE_CREDSTORE") != "" {
		return h.getFromCredSocket(serverURL)
	}

	buildID := os.Getenv("FINCH_BUILD_ID")
	if buildID == "" {
		return "", "", credentials.NewErrCredentialsNotFound()
	}

	credentialSocketPath := os.Getenv("FINCH_CREDENTIAL_SOCKET")
	if credentialSocketPath == "" {
		credentialSocketPath = CredentialSocketPath
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", credentialSocketPath)
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

	if resp.StatusCode != http.StatusOK {
		return "", "", credentials.NewErrCredentialsNotFound()
	}

	// Process successful response
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("error reading response: %v", err)
		return "", "", fmt.Errorf("error reading response: %v", err)
	}

	var errorResponse struct {
		Message string `json:"message,omitempty"`
	}
	if err := json.Unmarshal(responseBytes, &errorResponse); err == nil && errorResponse.Message != "" {
		return "", "", credentials.NewErrCredentialsNotFound()
	}

	var authConfig dockertypes.AuthConfig
	if err := json.Unmarshal(responseBytes, &authConfig); err != nil {
		return "", "", credentials.NewErrCredentialsNotFound()
	}

	return authConfig.Username, authConfig.Password, nil
}

func (h FinchCredentialHelper) getFromCredSocket(serverURL string) (string, string, error) {
	credentialSocketPath := os.Getenv("FINCH_CREDENTIAL_SOCKET")
	if credentialSocketPath == "" {
		// detect WSL; Windows Finch uses direct fs mount instead of port forwarding
		if strings.Contains(os.Getenv("PATH"), "/mnt/c") || os.Getenv("WSL_DISTRO_NAME") != "" {
			finchDir := os.Getenv("FINCH_DIR")
			if finchDir != "" {
				credentialSocketPath = filepath.Join(finchDir, "lima", "data", "finch", "sock", "creds.sock")
			} else {
				return "", "", fmt.Errorf("FINCH_DIR environment variable not set; cannot find mounted socket")
			}
		} else {
			// reverse port forwarded sock from mac.yaml
			credentialSocketPath = "/run/finch-user-sockets/creds.sock"
		}
	}

	// connect to socket
	conn, err := net.Dial("unix", credentialSocketPath)
	if err != nil {
		return "", "", credentials.NewErrCredentialsNotFound()
	}
	defer conn.Close()

	// simple sanitize to fight injection
	serverURL = strings.ReplaceAll(serverURL, "\n", "")
	serverURL = strings.ReplaceAll(serverURL, "\r", "")

	// send get command with URL through socket
	_, err = conn.Write([]byte("get\n" + serverURL + "\n"))
	if err != nil {
		return "", "", credentials.NewErrCredentialsNotFound()
	}

	// read response
	response := make([]byte, BufferSize)
	n, err := conn.Read(response)
	if err != nil {
		return "", "", credentials.NewErrCredentialsNotFound()
	}

	// parse response into credential struct (auth config breaking)
	var cred struct {
		ServerURL string `json:"ServerURL"`
		Username  string `json:"Username"`
		Secret    string `json:"Secret"`
	}
	if err := json.Unmarshal(response[:n], &cred); err != nil {
		return "", "", credentials.NewErrCredentialsNotFound()
	}

	// Return empty credentials if no credentials found
	if cred.Username == "" && cred.Secret == "" {
		return "", "", credentials.NewErrCredentialsNotFound()
	}

	return cred.Username, cred.Secret, nil
}

func main() {
	credentials.Serve(FinchCredentialHelper{})
}