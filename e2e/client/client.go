// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package client provides a client for communicating with finch-daemon using unix sockets.
package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// NewClient creates a new http client that connects to the finch-daemon server.
func NewClient(socketPath string) *http.Client {
	if socketPath == "" {
		panic("socketPath is empty")
	}
	// remove the prefix unix:// from the socket path
	socketPath = strings.TrimPrefix(socketPath, "unix://")

	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}
}

func ConvertToFinchUrl(version, relativeUrl string) string {
	if version == "" {
		return fmt.Sprintf("http://finch%s", relativeUrl)
	}
	return fmt.Sprintf("http://finch/%s%s", version, relativeUrl)
}
