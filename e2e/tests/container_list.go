// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"
	"github.com/runfinch/finch-daemon/e2e/client"
	"github.com/runfinch/finch-daemon/pkg/api/types"
)

// ContainerList tests the `GET containers/json` API.
func ContainerList(opt *option.Option) {
	Describe("list containers", func() {
		var (
			uClient                               *http.Client
			version                               string
			wantContainerName, wantContainerName2 string
		)
		BeforeEach(func() {
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
			wantContainerName = fmt.Sprintf("/%s", testContainerName)
			wantContainerName2 = fmt.Sprintf("/%s", testContainerName2)
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should list the running containers with no query parameter", func() {
			id := command.StdoutStr(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
			want := []types.ContainerListItem{
				{
					Id:    id[:12],
					Names: []string{wantContainerName},
				},
			}

			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got []types.ContainerListItem
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(len(got)).Should(Equal(2))
			got = filterContainerList(got)
			Expect(got).Should(ContainElements(want))
		})
		It("should list all the containers with all is true", func() {
			id1 := command.StdoutStr(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
			id2 := command.StdoutStr(opt, "run", "-d", "--name", testContainerName2, defaultImage)
			want := []types.ContainerListItem{
				{
					Id:    id1[:12],
					Names: []string{wantContainerName},
				},
				{
					Id:    id2[:12],
					Names: []string{wantContainerName2},
				},
			}

			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json?all=true"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got []types.ContainerListItem
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(len(got)).Should(Equal(3))
			got = filterContainerList(got)
			Expect(got).Should(ContainElements(want))
		})
		It("should list all the containers with all is true and limit is 1", func() {
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
			id2 := command.StdoutStr(opt, "run", "-d", "--name", testContainerName2, defaultImage)
			want := []types.ContainerListItem{
				{
					Id:    id2[:12],
					Names: []string{wantContainerName2},
				},
			}

			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json?all=true&limit=1"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got []types.ContainerListItem
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(len(got)).Should(Equal(1))
			got = filterContainerList(got)
			Expect(got).Should(ContainElements(want))
		})

		It("should list empty list with all is true and filters including exited containers but there is no exited container", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json?all=true&filters={\"status\":[\"exited\"]}"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got []types.ContainerListItem
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(len(got)).Should(Equal(0))
		})
		It("should list the running containers as normal with size is true", func() {
			id := command.StdoutStr(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
			want := []types.ContainerListItem{
				{
					Id:    id[:12],
					Names: []string{wantContainerName},
				},
			}

			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json?size=true"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got []types.ContainerListItem
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(len(got)).Should(Equal(2))
			got = filterContainerList(got)
			Expect(got).Should(ContainElements(want))
		})
		It("should list the running containers with all is true and filters including exited status", func() {
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
			id2 := command.StdoutStr(opt, "run", "-d", "--name", testContainerName2, defaultImage)
			want := []types.ContainerListItem{
				{
					Id:    id2[:12],
					Names: []string{wantContainerName2},
				},
			}

			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json?all=true&filters={\"status\":[\"exited\"]}"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got []types.ContainerListItem
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(len(got)).Should(Equal(1))
			got = filterContainerList(got)
			Expect(got).Should(ContainElements(want))
		})
		It("should list the running containers with filters including labels", func() {
			id := command.StdoutStr(opt, "run", "-d", "--name", testContainerName, "--label", "com.example.foo=bar", defaultImage, "sleep", "infinity")
			want := []types.ContainerListItem{
				{
					Id:    id[:12],
					Names: []string{wantContainerName},
				},
			}

			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json?all=true&filters={\"label\":[\"com.example.foo=bar\"]}"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got []types.ContainerListItem
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(len(got)).Should(Equal(1))
			got = filterContainerList(got)
			Expect(got).Should(ContainElements(want))
		})
		It("should list the running containers with filters including network", func() {
			command.Run(opt, "network", "create", testNetwork)
			id := command.StdoutStr(opt, "run", "-d", "--name", testContainerName, "--network", testNetwork, defaultImage, "sleep", "infinity")
			want := []types.ContainerListItem{
				{
					Id:    id[:12],
					Names: []string{wantContainerName},
				},
			}

			res, err := uClient.Get(client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/json?all=true&filters={\"network\":[\"%s\"]}", testNetwork)))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got []types.ContainerListItem
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(len(got)).Should(Equal(1))
			got = filterContainerList(got)
			Expect(got).Should(ContainElements(want))
		})
		It("should return 400 error when all parameter is invalid", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json?all=invalid"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			defer res.Body.Close()
			errorMsg := fmt.Sprintf("invalid query parameter \\\"all\\\": %s", fmt.Errorf("strconv.ParseBool: parsing \\\"invalid\\\": invalid syntax"))
			Expect(body).Should(MatchJSON(`{"message": "` + errorMsg + `"}`))
		})
		It("should return 400 error when limit parameter is invalid", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json?limit=invalid"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			defer res.Body.Close()
			errorMsg := fmt.Sprintf("invalid query parameter \\\"limit\\\": %s", fmt.Errorf("strconv.ParseInt: parsing \\\"invalid\\\": invalid syntax"))
			Expect(body).Should(MatchJSON(`{"message": "` + errorMsg + `"}`))
		})
		It("should return 400 error when size parameter is invalid", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json?size=invalid"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			defer res.Body.Close()
			errorMsg := fmt.Sprintf("invalid query parameter \\\"size\\\": %s", fmt.Errorf("strconv.ParseBool: parsing \\\"invalid\\\": invalid syntax"))
			Expect(body).Should(MatchJSON(`{"message": "` + errorMsg + `"}`))
		})
		It("should return 400 error when filters parameter is invalid", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json?filters=invalid"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			defer res.Body.Close()
			errorMsg := fmt.Sprintf("invalid query parameter \\\"filters\\\": %s", fmt.Errorf("invalid character 'i' looking for beginning of value"))
			Expect(body).Should(MatchJSON(`{"message": "` + errorMsg + `"}`))
		})
	})
}

// Checks that the other field for containers is non-null, then sets it to a zero value for comparison with dummy values
func filterContainerList(got []types.ContainerListItem) []types.ContainerListItem {
	filtered := []types.ContainerListItem{}
	for _, cont := range got {
		Expect(cont.CreatedAt).ShouldNot(BeZero())
		Expect(cont.Image).ShouldNot(BeEmpty())
		Expect(cont.Labels).ShouldNot(BeNil())
		Expect(cont.State).ShouldNot(BeEmpty())
		if cont.State != "exited" {
			Expect(cont.NetworkSettings.DefaultNetworkSettings.IPAddress).ShouldNot(BeEmpty())
		}
		cont.CreatedAt = 0
		cont.Image = ""
		cont.Labels = nil
		cont.NetworkSettings = nil
		cont.Mounts = nil
		cont.State = ""
		filtered = append(filtered, cont)
	}
	return filtered
}
