// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"
	"github.com/runfinch/finch-daemon/e2e/client"
)

func ImageBuild(opt *option.Option) {
	Describe("image build", func() {
		const (
			buildkit = 2
		)

		var (
			uClient *http.Client
			version string
		)

		BeforeEach(func() {
			command.RemoveImages(opt)
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
		})

		AfterEach(func() {
			command.RemoveAll(opt)
		})

		createDockerfile := func(dir, contents string) (string, func()) {
			file, err := os.CreateTemp(dir, "Dockerfile")
			Expect(err).Should(BeNil())

			_, err = file.Write([]byte(contents))
			Expect(err).Should(BeNil())

			return file.Name(), func() {
				err := os.Remove(file.Name())
				Expect(err).Should(BeNil())
			}
		}

		createBuildContext := func(dockerfileContents string) (string, func()) {
			buildContext, err := os.MkdirTemp("", "build-context.*")
			Expect(err).Should(BeNil())

			_, cleanupDockerfile := createDockerfile(buildContext, dockerfileContents)
			return buildContext, func() {
				cleanupDockerfile()

				err := os.RemoveAll(buildContext)
				Expect(err).Should(BeNil())
			}
		}

		createBuildArchive := func(pathToBuildContext string) (string, func()) {
			tarArchive, err := os.CreateTemp("", "archive-*.tar")
			Expect(err).Should(BeNil())
			defer tarArchive.Close()

			tarWriter := tar.NewWriter(tarArchive)
			defer tarWriter.Close()

			filepath.Walk(pathToBuildContext, func(path string, fileInfo os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
				if err != nil {
					return err
				}

				header.Name = pathToBuildContext
				if err := tarWriter.WriteHeader(header); err != nil {
					return err
				}

				if fileInfo.IsDir() {
					return nil
				}

				file, err := os.Open(pathToBuildContext)
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(tarWriter, file)
				return err
			})

			return tarArchive.Name(), func() {
				err := os.Remove(tarArchive.Name())
				Expect(err).Should(BeNil())
			}
		}

		It("should add extra hosts to /etc/hosts of the built container image", func() {
			dockerfileContents := `FROM public.ecr.aws/docker/library/alpine:latest
RUN ping -c 5 pong`

			pathToBuildContext, cleanupBuildContext := createBuildContext(dockerfileContents)
			defer cleanupBuildContext()

			pathToTarArchive, cleanupTarArchive := createBuildArchive(pathToBuildContext)
			defer cleanupTarArchive()

			tarArchive, err := os.Open(pathToTarArchive)
			Expect(err).Should(BeNil())
			defer tarArchive.Close()

			extraHosts := "[\"pong:127.0.0.1\"]"
			relativeURL := fmt.Sprintf("/build?version=%d&extrahosts=%s", buildkit, extraHosts)
			url := client.ConvertToFinchUrl(version, url.QueryEscape(relativeURL))

			resp, err := uClient.Post(url, "application/x-tar", tarArchive)
			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			waitForResponse(resp)
		})
	})
}
