// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import "github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"

// VolumesListResponse is the response object expected by GET /volumes
// https://docs.docker.com/engine/api/v1.40/#tag/Volume
type VolumesListResponse struct {
	Volumes []native.Volume `json:"Volumes"`
}
