// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"flag"
	"os"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/e2e/tests"
)

// Subject defines which CLI the tests are run against, defaults to \"nerdctl\" in the user's PATH.
var Subject = flag.String("subject", "nerdctl", `which CLI the tests are run against, defaults to "nerdctl" in the user's PATH.`)

func TestRun(t *testing.T) {
	if os.Getenv("TEST_E2E") != "1" {
		t.Skip("E2E tests skipped. Set TEST_E2E=1 to run these tests")
	}
	opt, _ := option.New([]string{*Subject, "--namespace", "finch"})

	ginkgo.SynchronizedBeforeSuite(func() []byte {
		tests.SetupLocalRegistry(opt)
		return nil
	}, func(bytes []byte) {})

	ginkgo.SynchronizedAfterSuite(func() {
		tests.CleanupLocalRegistry(opt)
		// clean up everything after the local registry is cleaned up
		command.RemoveAll(opt)
	}, func() {})

	const description = "Finch Daemon Functional test"
	ginkgo.Describe(description, func() {
		// functional test for container APIs
		tests.ContainerCreate(opt)
		tests.ContainerStart(opt)
		tests.ContainerStop(opt)
		tests.ContainerRestart(opt)
		tests.ContainerRemove(opt)
		tests.ContainerList(opt)
		tests.ContainerRename(opt)
		tests.ContainerStats(opt)
		tests.ContainerAttach(opt)
		tests.ContainerLogs(opt)
		tests.ContainerKill(opt)
		tests.ContainerInspect(opt)
		tests.ContainerWait(opt)
		tests.ContainerPause(opt)

		// functional test for volume APIs
		tests.VolumeList(opt)
		tests.VolumeInspect(opt)
		tests.VolumeRemove(opt)

		// functional test for network APIs
		tests.NetworkCreate(opt)
		tests.NetworkRemove(opt)
		tests.NetworkList(opt)
		tests.NetworkInspect(opt)

		// functional test for image APIs
		tests.ImageRemove(opt)
		tests.ImagePush(opt)
		tests.ImagePull(opt)

		// functional test for system api
		tests.SystemVersion(opt)
		tests.SystemEvents(opt)

		// functional test for distribution api
		tests.DistributionInspect(opt)
	})

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, description)
}
