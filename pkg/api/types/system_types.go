// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

// VersionInfo contains the response of /version api.
type VersionInfo struct {
	Platform struct {
		Name string
	}
	Version       string
	ApiVersion    string
	MinAPIVersion string
	GitCommit     string
	Os            string
	Arch          string
	KernelVersion string
	Experimental  bool
	BuildTime     string
	Components    []ComponentVersion
}

// ComponentVersion describes the version information for a specific component.
// From https://github.com/moby/moby/blob/v20.10.8/api/types/types.go#L112-L117
type ComponentVersion struct {
	Name    string
	Version string
	Details map[string]string `json:",omitempty"`
}
