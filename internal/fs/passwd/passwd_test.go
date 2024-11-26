// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package passwd

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var (
	content = []string{
		"root:x:0:0:Super User:/root:/bin/bash",
		"sshd:x:74:74:Privilege-separated SSH:/usr/share/empty.sshd:/usr/sbin/nologin",
		"user:x:1000:1000::/home/user:/bin/bash",
	}
	entries = []Entry{
		{Username: "root", UID: 0, GID: 0, Home: "/root", Shell: "/bin/bash"},
		{Username: "sshd", UID: 74, GID: 74, Home: "/usr/share/empty.sshd", Shell: "/usr/sbin/nologin"},
		{Username: "user", UID: 1000, GID: 1000, Home: "/home/user", Shell: "/bin/bash"},
	}
)

func TestPasswd(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "UnitTests - passwd parsing")
}

var _ = ginkgo.Describe("passwd", func() {
	ginkgo.DescribeTable("when parsing entries",
		func(line string, entry Entry) {
			parsed, err := parse(line)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(parsed).To(gomega.Equal(entry))
		},
		ginkgo.Entry("should correctly parse root", content[0], entries[0]),
		ginkgo.Entry("should correctly parse a daemon user", content[1], entries[1]),
		ginkgo.Entry("should correctly parse a regular user", content[2], entries[2]),
	)

	ginkgo.DescribeTable("when walking passwd file", func(match bool, expectedCount int) {
		i := 0
		passwdReader := bytes.NewReader([]byte(strings.Join(content, "\n")))
		open = func() (io.ReadCloser, error) {
			return io.NopCloser(passwdReader), nil
		}
		Walk(func(e Entry) bool {
			gomega.Expect(e).To(gomega.Equal(entries[i]))
			i++
			return !match
		})
		gomega.Expect(i).To(gomega.Equal(expectedCount))
	},
		ginkgo.Entry("should see all entries if no matches", false, len(entries)),
		ginkgo.Entry("should see 1 entry if first matches", true, 1),
	)
},
)
