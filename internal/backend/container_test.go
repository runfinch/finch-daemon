// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetHookHelperBinary", func() {
	var w *NerdctlWrapper

	BeforeEach(func() {
		w = &NerdctlWrapper{}
	})

	It("returns the resolved path on first call", func() {
		w.lookPath = func(name string) (string, error) {
			Expect(name).To(Equal("finch-hook"))
			return "/usr/local/bin/finch-hook", nil
		}

		path, err := w.GetHookHelperBinary()
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal("/usr/local/bin/finch-hook"))
	})

	It("caches the result and calls lookPath exactly once on repeated calls", func() {
		callCount := 0
		w.lookPath = func(name string) (string, error) {
			callCount++
			return "/usr/local/bin/finch-hook", nil
		}

		first, err := w.GetHookHelperBinary()
		Expect(err).NotTo(HaveOccurred())

		second, err := w.GetHookHelperBinary()
		Expect(err).NotTo(HaveOccurred())

		Expect(first).To(Equal(second))
		Expect(callCount).To(Equal(1), "lookPath should be called exactly once; cached on second call")
	})

	It("returns an error containing 'finch-hook binary not found in PATH' when lookPath fails", func() {
		w.lookPath = func(name string) (string, error) {
			return "", errors.New("not found")
		}

		_, err := w.GetHookHelperBinary()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("finch-hook binary not found in PATH"))
	})

	It("returns the pre-set path without calling lookPath when hookHelperExe is already set", func() {
		w.hookHelperExe = "/custom/path/finch-hook"
		w.lookPath = func(name string) (string, error) {
			Fail("lookPath should not be called when hookHelperExe is already cached")
			return "", nil
		}

		path, err := w.GetHookHelperBinary()
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal("/custom/path/finch-hook"))
	})
})
