// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

/*
ImageSummary models a single item in the list response to /images/json in the
Docker API.
https://docs.docker.com/engine/api/v1.40/#operation/ImageList
*/
type ImageSummary struct {
	ID          string `json:"Id"`
	RepoTags    []string
	RepoDigests []string
	Created     int64
	Size        int64
}

// PushResult contains the tag, manifest digest, and manifest size from the
// push. It's used to signal this information to the trust code in the client
// so it can sign the manifest if necessary.
// From https://github.com/moby/moby/blob/v24.0.2/api/types/types.go#L765-L772
type PushResult struct {
	Tag    string
	Digest string
	Size   int
}
