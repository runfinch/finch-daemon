// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

// BuildResult contains the image id of a successful build
// From https://github.com/moby/moby/blob/v24.0.2/api/types/types.go#L774-L777
type BuildResult struct {
	ID string
}
