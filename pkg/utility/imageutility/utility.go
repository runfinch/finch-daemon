// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package imageutility

import (
	"strings"

	"github.com/distribution/reference"
)

const (
	defaultTag      = "latest"
	tagDigestPrefix = "sha256:"
	eventType       = "image"
)

func Canonicalize(name, tag string) (string, error) {
	if name != "" {
		if strings.HasPrefix(tag, tagDigestPrefix) {
			name += "@" + tag
		} else if tag != "" {
			name += ":" + tag
		}
	} else {
		name = tag
	}
	ref, err := reference.ParseAnyReference(name)
	if err != nil {
		return "", err
	}
	if named, ok := ref.(reference.Named); ok && refNeedsTag(ref) {
		tagged, err := reference.WithTag(named, defaultTag)
		if err == nil {
			ref = tagged
		}
	}
	return ref.String(), nil
}

func refNeedsTag(ref reference.Reference) bool {
	_, tagged := ref.(reference.Tagged)
	_, digested := ref.(reference.Digested)
	return !(tagged || digested)
}
