// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"flag"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/e2e/tests"
	"github.com/runfinch/finch-daemon/e2e/util"
)

// Subject defines which CLI the tests are run against, defaults to \"nerdctl\" in the user's PATH.
var Subject = flag.String("subject", "nerdctl", `which CLI the tests are run against, defaults to "nerdctl" in the user's PATH.`)
var SubjectPrefix = flag.String("daemon-context-subject-prefix", "", `A string which prefixes the command the tests are run against, defaults to "". This string will be split by spaces.`)
var PrefixedSubjectEnv = flag.String("daemon-context-subject-env", "", `Environment to add when running a prefixed subject, in the form of a string like "EXAMPLE=foo EXAMPLE2=bar"`)

func TestRun(t *testing.T) {
	if os.Getenv("TEST_E2E") != "1" {
		t.Skip("E2E tests skipped. Set TEST_E2E=1 to run these tests")
	}

	if err := parseTestFlags(); err != nil {
		log.Println("failed to parse go test flags", err)
		os.Exit(1)
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

	var pOpt = option.New
	if *SubjectPrefix != "" {
		var modifiers []option.Modifier
		if *PrefixedSubjectEnv != "" {
			modifiers = append(modifiers, option.Env(strings.Split(*PrefixedSubjectEnv, " ")))
		}
		pOpt = util.WrappedOption(strings.Split(*SubjectPrefix, " "), modifiers...)
	}

	const description = "Finch Daemon Functional test"
	ginkgo.Describe(description, func() {
		// functional test for container APIs
		tests.ContainerCreate(opt, pOpt)
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
		tests.ContainerUnpause(opt)

		// functional test for volume APIs
		tests.VolumeList(opt)
		tests.VolumeInspect(opt)
		tests.VolumeRemove(opt)

		// functional test for network APIs
		tests.NetworkCreate(opt, pOpt)
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

// parseTestFlags parses go test flags because pflag package ignores flags with '-test.' prefix
// Related issues:
// https://github.com/spf13/pflag/issues/63
// https://github.com/spf13/pflag/issues/238
func parseTestFlags() error {
	var testFlags []string
	for _, f := range os.Args[1:] {
		if strings.HasPrefix(f, "-test.") {
			testFlags = append(testFlags, f)
		}
	}
	return flag.CommandLine.Parse(testFlags)
}
