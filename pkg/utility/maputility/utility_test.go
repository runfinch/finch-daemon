// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package maputility

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMapUtility(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Map Utility Functions")
}

var _ = Describe("Map Utility", func() {
	When("flattening an empty map", func() {
		It("should return an empty array", func() {
			empty := make(map[string]string, 0)
			Expect(Flatten(empty, KeyEqualsValueFormat)).Should(BeEmpty())
		})
	})

	When("flattening a map with entries", func() {
		It("should return an array with key=value format", func() {
			mapWithEntries := map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}
			actual := Flatten(mapWithEntries, KeyEqualsValueFormat)
			Expect(actual).ShouldNot(BeEmpty())
			Expect(actual).To(ContainElements("key1=value1", "key2=value2", "key3=value3"))
		})
	})
})
