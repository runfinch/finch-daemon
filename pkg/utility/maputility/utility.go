// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package maputility

import "fmt"

type MapFormatFunction func(string, string) string

// KeyEqualsValueFormat is a formatting function for formatting a map entry
// like { "key": "value"} into "key=value" when flattening a map.
func KeyEqualsValueFormat(key string, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

// Flatten reduces a key-value map into a string array using the provided
// formatting function.
func Flatten(kvMap map[string]string, format MapFormatFunction) []string {
	return reduce(kvMap, []string{}, format)
}

func reduce(collection map[string]string, initial []string, reduce func(string, string) string) []string {
	accumulator := initial

	for k, v := range collection {
		accumulator = append(accumulator, reduce(k, v))
	}

	return accumulator
}
