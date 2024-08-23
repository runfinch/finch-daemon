// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package version provides the global default latest API server version for the project
package version

var (
	// Version and GitCommit value is set from the make file.
	Version           string
	GitCommit         string
	DefaultApiVersion = "1.43"
	MinimumApiVersion = "1.35"
)
