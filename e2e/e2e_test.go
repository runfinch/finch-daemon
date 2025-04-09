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

const (
	defaultNamespace   = "finch"
	testE2EEnv         = "TEST_E2E"
	middlewareE2EEnv   = "MIDDLEWARE_E2E"
	opaTestDescription = "Finch Daemon OPA E2E Tests"
	e2eTestDescription = "Finch Daemon Functional test"
)

var (
	Subject            = flag.String("subject", "nerdctl", `which CLI the tests are run against, defaults to "nerdctl" in the user's PATH.`)
	SubjectPrefix      = flag.String("daemon-context-subject-prefix", "", `A string which prefixes the command the tests are run against, defaults to "". This string will be split by spaces.`)
	PrefixedSubjectEnv = flag.String("daemon-context-subject-env", "", `Environment to add when running a prefixed subject, in the form of a string like "EXAMPLE=foo EXAMPLE2=bar"`)
)

func TestRun(t *testing.T) {
	switch {
	case os.Getenv(middlewareE2EEnv) == "1":
		runOPATests(t)
	case os.Getenv(testE2EEnv) == "1":
		runE2ETests(t)
	default:
		t.Skip("E2E tests skipped. Set TEST_E2E=1 to run regular E2E tests or MIDDLEWARE_E2E=1 to run OPA middleware tests")
	}
}

func createTestOption() (*option.Option, error) {
	return option.New([]string{*Subject, "--namespace", defaultNamespace})
}

func setupTestSuite(opt *option.Option) {
	ginkgo.SynchronizedBeforeSuite(func() []byte {
		tests.SetupLocalRegistry(opt)
		return nil
	}, func(bytes []byte) {})

	ginkgo.SynchronizedAfterSuite(func() {
		tests.CleanupLocalRegistry(opt)
		command.RemoveAll(opt)
	}, func() {})
}

func runOPATests(t *testing.T) {
	if err := parseTestFlags(); err != nil {
		log.Fatal("failed to parse go test flags:", err)
	}

	opt, err := createTestOption()
	if err != nil {
		log.Fatal("failed to create test option:", err)
	}

	setupTestSuite(opt)

	ginkgo.Describe(opaTestDescription, func() {
		tests.OpaMiddlewareTest(opt)
	})

	runTests(t, opaTestDescription)
}

func runE2ETests(t *testing.T) {
	if err := parseTestFlags(); err != nil {
		log.Fatal("failed to parse go test flags:", err)
	}

	opt, err := createTestOption()
	if err != nil {
		log.Fatal("failed to create test option:", err)
	}

	setupTestSuite(opt)

	pOpt := createPrefixedOption()

	ginkgo.Describe(e2eTestDescription, func() {
		runContainerTests(opt)
		runVolumeTests(opt)
		runNetworkTests(opt, pOpt)
		runImageTests(opt)
		runSystemTests(opt)
		runDistributionTests(opt)
	})

	runTests(t, e2eTestDescription)
}

func createPrefixedOption() func([]string, ...option.Modifier) (*option.Option, error) {
	if *SubjectPrefix == "" {
		return option.New
	}

	var modifiers []option.Modifier
	if *PrefixedSubjectEnv != "" {
		modifiers = append(modifiers, option.Env(strings.Split(*PrefixedSubjectEnv, " ")))
	}
	return util.WrappedOption(strings.Split(*SubjectPrefix, " "), modifiers...)
}

func runTests(t *testing.T, description string) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, description)
}

// functional test for container APIs.
func runContainerTests(opt *option.Option) {
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
}

// functional test for volume APIs.
func runVolumeTests(opt *option.Option) {
	tests.VolumeList(opt)
	tests.VolumeInspect(opt)
	tests.VolumeRemove(opt)
}

// functional test for network APIs.
func runNetworkTests(opt *option.Option, pOpt func([]string, ...option.Modifier) (*option.Option, error)) {
	tests.NetworkCreate(opt, pOpt)
	tests.NetworkRemove(opt)
	tests.NetworkList(opt)
	tests.NetworkInspect(opt)
}

// functional test for image APIs.
func runImageTests(opt *option.Option) {
	tests.ImageRemove(opt)
	tests.ImagePush(opt)
	tests.ImagePull(opt)
}

// .
func runSystemTests(opt *option.Option) {
	tests.SystemVersion(opt)
	tests.SystemEvents(opt)
}

// functional test for distribution api.
func runDistributionTests(opt *option.Option) {
	tests.DistributionInspect(opt)
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
