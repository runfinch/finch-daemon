// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"

	gocni "github.com/containerd/go-cni"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/containerd/nerdctl/v2/pkg/defaults"
	"github.com/docker/go-connections/nat"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Container Create API ", func() {
	var (
		mockCtrl     *gomock.Controller
		logger       *mocks_logger.Logger
		service      *mocks_container.MockService
		createOpt    types.ContainerCreateOptions
		netOpt       types.NetworkOptions
		cid          string
		jsonResponse interface{}
		h            *handler
		rr           *httptest.ResponseRecorder
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_container.NewMockService(mockCtrl)
		c := config.Config{}
		createOpt = getDefaultCreateOpt(c)
		netOpt = getDefaultNetOpt()
		cid = "123"
		jsonResponse = `{"Id": "123"}`
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
	})
	Context("handler", func() {
		It("should return 201 as success response", func() {
			body := []byte(`{"Image": "test-image"}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// service mock returns container id and nil error upon success.
			service.EXPECT().Create(gomock.Any(), "test-image", gomock.Nil(), equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set the Cmd argument", func() {
			body := []byte(`{
				"Image": "test-image",
				"Cmd": ["echo", "hello world"]
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			service.EXPECT().Create(gomock.Any(), "test-image", []string{"echo", "hello world"}, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should handle valid port mappings", func() {
			body := []byte(`{
				"Image": "test-image",
				"ExposedPorts": {"8000/tcp": {}, "9000/udp": {}},
				"HostConfig": {
					"PortBindings": {
						"8000/tcp": [{"HostIp": "", "HostPort": "8001"}],
						"9000/udp": [{"HostIp": "127.0.0.1", "HostPort": "9001"}]
					}
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// define expected go-cni port mappings and network settings
			portMaps := []gocni.PortMapping{
				{
					HostPort:      8001,
					ContainerPort: 8000,
					Protocol:      "tcp",
					HostIP:        "",
				},
				{
					HostPort:      9001,
					ContainerPort: 9000,
					Protocol:      "udp",
					HostIP:        "127.0.0.1",
				},
			}

			// port mappings can be in any order
			netOpt1 := types.NetworkOptions{
				Hostname:             netOpt.Hostname,
				NetworkSlice:         netOpt.NetworkSlice,
				DNSResolvConfOptions: netOpt.DNSResolvConfOptions,
				PortMappings:         portMaps,
			}
			netOpt2 := types.NetworkOptions{
				Hostname:             netOpt.Hostname,
				NetworkSlice:         netOpt.NetworkSlice,
				DNSResolvConfOptions: netOpt.DNSResolvConfOptions,
				PortMappings:         []gocni.PortMapping{portMaps[1], portMaps[0]},
			}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), anyOf(netOpt1, netOpt2)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should return 400 for invalid port mappings", func() {
			body := []byte(`{
				"Image": "test-image",
				"ExposedPorts": {"8000/tcp": {}},
				"HostConfig": {
					"PortBindings": {
						"8000/tcp": [{"HostIp": "", "HostPort": ""}],
					}
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// handler should return bad request message with 400 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should set the default network mode to bridge", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"NetworkMode": "default"
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set the specified network mode", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"NetworkMode": "net1"
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// define expected network mode
			netOpt.NetworkSlice = []string{"net1"}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set container name and platform parameters", func() {
			body := []byte(`{"Image": "test-image"}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create?name=test-cont&platform=arm64", bytes.NewReader(body))

			// expected name and platform parameters
			createOpt.Name = "test-cont"
			createOpt.Platform = "arm64"

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set specified create options", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"AutoRemove": true,
					"Memory": 209715200,
					"RestartPolicy": {
						"Name": "on-failure",
						"MaximumRetryCount": 0
					}
				},
				"User": "test-user",
				"Env": ["VARIABLE1=1", "VAR2=var2"],
				"WorkingDir": "/test-dir",
				"Entrypoint": ["echo", "hello"],
				"StopSignal": "SIGINT",
				"StopTimeout": 500
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.Rm = true
			createOpt.Restart = "on-failure"
			createOpt.User = "test-user"
			createOpt.Env = []string{"VARIABLE1=1", "VAR2=var2"}
			createOpt.Workdir = "/test-dir"
			createOpt.Entrypoint = []string{"echo", "hello"}
			createOpt.EntrypointChanged = true
			createOpt.StopSignal = "SIGINT"
			createOpt.StopTimeout = 500
			createOpt.Memory = "209715200"

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set specified create options for resources", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"Memory": 209715200,
					"CPUShares": 1
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.Memory = "209715200"
			createOpt.CPUShares = 1

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set specified create options for logging", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"LogConfig": {
						"Type": "json-file",
						"Config": {
							"key": "value"
						}
					}
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.LogDriver = "json-file"
			createOpt.LogOpt = []string{"key=value"}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set specified network options", func() {
			body := []byte(`{
				"Image": "test-image",
				"Hostname": "test-host",
				"HostConfig": {
					"DNS": ["8.8.8.8"],
					"DNSOptions": ["test-opt"],
					"DNSSearch": ["test.com"],
					"ExtraHosts": ["test-host:127.0.0.1"]
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected network options
			netOpt.Hostname = "test-host"
			netOpt.DNSServers = []string{"8.8.8.8"}
			netOpt.DNSResolvConfOptions = []string{"test-opt"}
			netOpt.DNSSearchDomains = []string{"test.com"}
			netOpt.AddHost = []string{"test-host:127.0.0.1"}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set specified volume mounts", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"Binds": ["/tmp/workdir:/workdir:ro,delegated", "test-vol1:/mnt/test-vol1", "test-vol2"]
				},
				"Volumes": {
					"test-vol3": {},
					"/workdir": {}
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected volume options
			createOpt.Volume = []string{
				"/tmp/workdir:/workdir:ro,delegated",
				"test-vol1:/mnt/test-vol1",
				"test-vol2",
				"test-vol3",
			}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set CPUPeriod create options for resources", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"CpuPeriod": 100000
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.CPUPeriod = 100000

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set CpuQuota create options for resources", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"CpuQuota": 50000
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.CPUQuota = 50000
			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set CpuQuota to -1 by default", func() {
			body := []byte(`{
				"Image": "test-image"
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.CPUQuota = -1
			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set CpuSet create options for resources", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"CpusetCpus": "0,1",
					"CpusetMems": "0,3"
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.CPUSetCPUs = "0,1"
			createOpt.CPUSetMems = "0,3"
			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set MemoryReservation, MemorySwap and MemorySwappiness create options for resources", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"MemoryReservation": 209710,
					"MemorySwap": 514288000,
					"MemorySwappiness": 25
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.MemoryReservation = "209710"
			createOpt.MemorySwap = "514288000"
			createOpt.MemorySwappiness64 = 25
			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set CapDrop option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"CapDrop": ["MKNOD"]
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.CapDrop = []string{"MKNOD"}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set GroupAdd option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"GroupAdd": ["someGroup"]
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.GroupAdd = []string{"someGroup"}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set Privileged option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"Privileged": true
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.Privileged = true

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set Ulimit option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"Ulimits": [{"Name": "nofile", "Soft": 1024, "Hard": 2048},{"Name": "nproc", "Soft": 1024, "Hard": 4048}]
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.Ulimit = []string{"nofile=1024:2048", "nproc=1024:4048"}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set PidLimit option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"PidsLimit": 200
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.PidsLimit = 200

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set ContainerIdFile option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"ContainerIDFile": "/lib/example.txt"
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.CidFile = "/lib/example.txt"

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should return 404 if the image was not found", func() {
			body := []byte(`{"Image": "test-image"}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				"", errdefs.NewNotFound(errors.New("error message")))

			// handler should return error message with 404 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
		})

		It("should return 400 if the inputs are invalid", func() {
			body := []byte(`{"Image": "test-image"}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				"", errdefs.NewInvalidFormat(errors.New("error message")))

			// handler should return error message with 400 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should return 409 if the container already exists", func() {
			body := []byte(`{"Image": "test-image"}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				"", errdefs.NewConflict(errors.New("error message")))

			// handler should return error message with 409 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusConflict))
		})

		It("should return 500 for internal failures", func() {
			body := []byte(`{"Image": "test-image"}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				"", errors.New("error message"))

			// handler should return error message with 500 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})

		It("should return 400 Bad Request for container attach stdin during create", func() {
			body := []byte(`{"AttachStdin": true}`)
			req, err := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))
			Expect(err).ShouldNot(HaveOccurred())

			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body.String()).Should(ContainSubstring("not supported"))
		})

		It("should return 400 Bad Request for invalid port mappings during create", func() {
			body := []byte(`{"HostConfig": {"PortBindings": {"22/tcp": [{"HostPort": "Twenty-Two"}]}}}`)
			req, err := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))
			Expect(err).ShouldNot(HaveOccurred())

			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body.String()).Should(ContainSubstring("failed to parse"))
		})

		It("should set specified NetworkDisabled setting", func() {
			body := []byte(`{
				"Image": "test-image",
				"NetworkDisabled": true
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected network options
			netOpt.NetworkSlice = []string{"none"}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set the MACAddress to a user specified value", func() {
			body := []byte(`{
				"Image": "test-image",
				"MacAddress": "12:34:56:78:9a:bc"
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected network options
			netOpt.MACAddress = "12:34:56:78:9a:bc"
			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set the OomKillDisable option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"OomKillDisable": true
					}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected network options
			createOpt.OomKillDisable = true
			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})
		It("should set the BlkioWeight to a user specified value", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"BlkioWeight": 300
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected network options
			createOpt.BlkioWeight = 300

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set blkio device settings", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"BlkioWeightDevice": [
						{
							"Path": "/dev/sda",
							"Weight": 400
						}
					],
					"BlkioDeviceReadBps": [
						{
							"Path": "/dev/sda",
							"Rate": 1048576
						}
					],
					"BlkioDeviceWriteBps": [
						{
							"Path": "/dev/sda",
							"Rate": 2097152
						}
					],
					"BlkioDeviceReadIOps": [
						{
							"Path": "/dev/sda",
							"Rate": 1000
						}
					],
					"BlkioDeviceWriteIOps": [
						{
							"Path": "/dev/sda",
							"Rate": 2000
						}
					]
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options with blkio settings as strings
			createOpt.BlkioWeightDevice = []string{
				"/dev/sda:400",
			}
			createOpt.BlkioDeviceReadBps = []string{
				"/dev/sda:1048576",
			}
			createOpt.BlkioDeviceWriteBps = []string{
				"/dev/sda:2097152",
			}
			createOpt.BlkioDeviceReadIOps = []string{
				"/dev/sda:1000",
			}
			createOpt.BlkioDeviceWriteIOps = []string{
				"/dev/sda:2000",
			}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set VolumesFrom option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"VolumesFrom": [ "parent", "other:ro"]
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			createOpt.VolumesFrom = []string{"parent", "other:ro"}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set Tmpfs and UTSMode option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"Tmpfs": { "/run": "rw,noexec,nosuid,size=65536k" },
					"UTSMode": "host"
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.Tmpfs = []string{"/run:rw,noexec,nosuid,size=65536k"}
			netOpt.UTSNamespace = "host"

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set PidMode option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"PidMode": "host"
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.Pid = "host"

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set IPC option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"IpcMode": "host"
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.IPC = "host"

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set ShmSize, Sysctl and Runtime option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"Sysctls": { "net.ipv4.ip_forward": "1" },
					"ShmSize": 302348,
					"Runtime": "crun"
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.ShmSize = "302348"
			createOpt.Sysctl = []string{"net.ipv4.ip_forward=1"}
			createOpt.Runtime = "crun"

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should set ReadonlyRootfs and SecurityOpt option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"ReadonlyRootfs": true,
					"SecurityOpt": [ "seccomp=/path/to/custom_seccomp.json", "apparmor=unconfined"]
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.ReadOnly = true
			createOpt.SecurityOpt = []string{"seccomp=/path/to/custom_seccomp.json", "apparmor=unconfined"}

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should reject invalid device paths", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"Devices": [
						{
							"PathOnHost": "/dev/nonexistent",
							"PathInContainer": "/dev/nonexistent",
							"CgroupPermissions": "rwm"
						}
					]
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body.String()).Should(ContainSubstring(`{"message":"error gathering device information while adding custom device \"/dev/nonexistent\": lstat /dev/nonexistent: no such file or directory"}`))
		})

		It("should reject empty device paths", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"Devices": [
						{
							"PathOnHost": "",
							"PathInContainer": "/dev/null",
							"CgroupPermissions": "rwm"
						}
					]
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body.String()).Should(ContainSubstring(`{"message":"error gathering device information while adding custom device"}`))
		})

		It("should reject non-device paths", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"Devices": [
						{
							"PathOnHost": "/etc/hosts",
							"PathInContainer": "/dev/null",
							"CgroupPermissions": "rwm"
						}
					]
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body.String()).Should(ContainSubstring(`{"message":"error gathering device information while adding custom device \"/etc/hosts\": not a device"}`))
		})

		It("should set CgroupnsMode option", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"CgroupnsMode": "host"
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// expected create options
			createOpt.Cgroupns = "host"

			service.EXPECT().Create(gomock.Any(), "test-image", nil, equalTo(createOpt), equalTo(netOpt)).Return(
				cid, nil)

			// handler should return success message with 201 status code.
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(jsonResponse))
		})

		It("should reject invalid CgroupnsMode values", func() {
			body := []byte(`{
				"Image": "test-image",
				"HostConfig": {
					"CgroupnsMode": "invalid"
				}
			}`)
			req, _ := http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))

			// handler should return bad request with error message
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body.String()).Should(ContainSubstring("invalid cgroup namespace mode"))
		})

		Context("translate port mappings", func() {
			It("should return empty if port mappings is nil", func() {
				Expect(translatePortMappings(nil)).Should(BeEmpty())
			})

			It("should return an error if port map binding is invalid", func() {
				portMappings := nat.PortMap{
					"80/tcp": {
						nat.PortBinding{
							HostIP:   "127.0.0.1",
							HostPort: "invalid-port-number",
						},
					},
				}
				cniPortMappings, err := translatePortMappings(portMappings)
				Expect(err).Should(HaveOccurred())
				Expect(cniPortMappings).Should(BeEmpty())
			})

			It("should return an error if container port is invalid", func() {
				portMappings := nat.PortMap{
					"invalid-port-number/tcp": {
						nat.PortBinding{
							HostIP:   "127.0.0.1",
							HostPort: "300",
						},
					},
				}
				cniPortMappings, err := translatePortMappings(portMappings)
				Expect(err).Should(HaveOccurred())
				Expect(cniPortMappings).Should(BeEmpty())
			})

			It("should return the expected port mappings", func() {
				expected := []gocni.PortMapping{
					{
						HostPort:      300,
						ContainerPort: 80,
						Protocol:      "tcp",
						HostIP:        "127.0.0.1",
					},
					{
						HostPort:      42,
						ContainerPort: 8080,
						Protocol:      "tcp",
						HostIP:        "127.0.0.1",
					},
				}
				portMappings := nat.PortMap{
					"80/tcp": {
						nat.PortBinding{
							HostIP:   "127.0.0.1",
							HostPort: "300",
						},
					},
					"8080/tcp": {
						nat.PortBinding{
							HostIP:   "127.0.0.1",
							HostPort: "42",
						},
					},
				}
				cniPortMappings, err := translatePortMappings(portMappings)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cniPortMappings).ShouldNot(BeEmpty())
				Expect(cniPortMappings).Should(ContainElements(expected))
			})
		})

		Context("translate tmpfs", func() {
			It("should return nil for nil input", func() {
				Expect(translateTmpfs(nil)).Should(BeNil())
			})

			It("should return empty slice for empty map", func() {
				Expect(translateTmpfs(map[string]string{})).Should(BeEmpty())
			})

			It("should handle single tmpfs mount with options", func() {
				input := map[string]string{
					"/run": "rw,noexec,nosuid,size=65536k",
				}
				expected := []string{"/run:rw,noexec,nosuid,size=65536k"}
				Expect(translateTmpfs(input)).Should(Equal(expected))
			})

			It("should handle multiple tmpfs mounts with different options", func() {
				input := map[string]string{
					"/run": "rw,noexec,nosuid,size=65536k",
					"/tmp": "rw,exec,size=32768k",
					"/var": "",
				}
				result := translateTmpfs(input)
				Expect(result).Should(ConsistOf(
					"/run:rw,noexec,nosuid,size=65536k",
					"/tmp:rw,exec,size=32768k",
					"/var",
				))
			})

			It("should handle tmpfs mount without options", func() {
				input := map[string]string{
					"/run": "",
				}
				expected := []string{"/run"}
				Expect(translateTmpfs(input)).Should(Equal(expected))
			})
		})

		Context("translate annotations", func() {
			It("should return nil for nil input", func() {
				Expect(translateAnnotations(nil)).Should(BeNil())
			})

			It("should return empty slice for empty map", func() {
				Expect(translateAnnotations(map[string]string{})).Should(BeEmpty())
			})

			It("should handle single annotation", func() {
				input := map[string]string{
					"com.example.key": "value1",
				}
				expected := []string{"com.example.key=value1"}
				Expect(translateAnnotations(input)).Should(Equal(expected))
			})

			It("should handle multiple annotations", func() {
				input := map[string]string{
					"com.example.key1":  "value1",
					"com.example.key2":  "value2",
					"io.containerd.key": "value3",
				}
				result := translateAnnotations(input)
				Expect(result).Should(ConsistOf(
					"com.example.key1=value1",
					"com.example.key2=value2",
					"io.containerd.key=value3",
				))
			})
		})
	})
})

// define default container create options.
func getDefaultCreateOpt(conf config.Config) types.ContainerCreateOptions {
	globalOpt := types.GlobalCommandOptions{
		Debug:            conf.Debug,
		DebugFull:        conf.DebugFull,
		Address:          conf.Address,
		Namespace:        conf.Namespace,
		Snapshotter:      conf.Snapshotter,
		CNIPath:          conf.CNIPath,
		CNINetConfPath:   conf.CNINetConfPath,
		DataRoot:         conf.DataRoot,
		CgroupManager:    conf.CgroupManager,
		InsecureRegistry: conf.InsecureRegistry,
		HostsDir:         conf.HostsDir,
		Experimental:     conf.Experimental,
		HostGatewayIP:    conf.HostGatewayIP,
	}
	return types.ContainerCreateOptions{
		Stdout:   nil,
		Stderr:   nil,
		GOptions: globalOpt,

		// #region for basic flags
		Interactive: false,     // TODO: update this after attach supports STDIN
		TTY:         false,     // TODO: update this after attach supports STDIN
		Detach:      true,      // TODO: current implementation of create does not support AttachStdin, AttachStdout, and AttachStderr flags
		Restart:     "no",      // Docker API default.
		Rm:          false,     // Automatically remove container upon exit
		Pull:        "missing", // nerdctl default.
		StopSignal:  "SIGTERM",
		StopTimeout: 10,
		// #endregion

		// #region for platform flags
		Platform: "", // target platform
		// #endregion

		// #region for isolation flags
		Isolation: "default", // nerdctl default.
		// #endregion

		// #region for resource flags
		CPUQuota:           -1,                      // nerdctl default.
		MemorySwappiness64: -1,                      // nerdctl default.
		PidsLimit:          -1,                      // nerdctl default.
		Cgroupns:           defaults.CgroupnsMode(), // nerdctl default.
		Ulimit:             []string{},

		BlkioWeightDevice:    []string{},
		BlkioDeviceReadBps:   []string{},
		BlkioDeviceWriteBps:  []string{},
		BlkioDeviceReadIOps:  []string{},
		BlkioDeviceWriteIOps: []string{},
		Device:               []string{},
		// #endregion

		// #region for user flags
		User: "",
		// #endregion

		// #region for security flags
		SecurityOpt: []string{}, // nerdctl default.
		CapAdd:      []string{}, // nerdctl default.
		CapDrop:     []string{}, // nerdctl default.
		Privileged:  false,
		GroupAdd:    []string{}, // nerdctl default.
		// #endregion

		// #region for runtime flags
		Runtime: defaults.Runtime, // nerdctl default.
		Sysctl:  []string{},
		// #endregion

		// #region for volume flags
		Volume:      nil,
		VolumesFrom: []string{},
		Tmpfs:       []string{},
		// #endregion

		// #region for env flags
		Env:               []string{},
		Workdir:           "",
		Entrypoint:        nil,
		EntrypointChanged: false,
		// #endregion

		// #region for metadata flags
		Name:  "",         // container name
		Label: []string{}, // container labels
		// #endregion

		// #region for logging flags
		LogDriver: "json-file", // nerdctl default.
		LogOpt:    []string{},
		// #endregion

		// #region for image pull and verify types
		ImagePullOpt: types.ImagePullOptions{
			GOptions:      globalOpt,
			VerifyOptions: types.ImageVerifyOptions{Provider: "none"},
			IPFSAddress:   "",
			Stdout:        nil,
			Stderr:        nil,
		},
		// #endregion
	}
}

// define default network types.
func getDefaultNetOpt() types.NetworkOptions {
	return types.NetworkOptions{
		Hostname:             "",
		NetworkSlice:         []string{"bridge"}, // nerdctl default.
		DNSResolvConfOptions: []string{},         // nerdctl default.
		PortMappings:         []gocni.PortMapping{},
	}
}

// anyOfMatcher is a gomock matcher that returns true if the object is contained in an array slice.
type anyOfMatcher struct {
	slice []interface{}
}

func anyOf(elements ...interface{}) *anyOfMatcher {
	return &anyOfMatcher{elements}
}

func (a *anyOfMatcher) Matches(x interface{}) bool {
	for _, element := range a.slice {
		if reflect.DeepEqual(element, x) {
			return true
		}
	}
	return false
}

func (a *anyOfMatcher) String() string {
	return fmt.Sprintf("one of the elements in slice: %v", a.slice)
}

// equalToMatcher is a gomock matcher similar to gomock.Eq(), but it prints specific fields upon mismatch.
// This is useful for comparing large structs.
type equalToMatcher struct {
	obj        interface{}
	mismatches []string
}

func equalTo(object interface{}) *equalToMatcher {
	return &equalToMatcher{
		obj:        object,
		mismatches: []string{},
	}
}

func (e *equalToMatcher) Matches(x interface{}) bool {
	e.mismatches = []string{}
	v1 := reflect.ValueOf(e.obj)
	v2 := reflect.ValueOf(x)
	t := reflect.TypeOf(e.obj)
	for i := 0; i < v1.NumField(); i++ {
		f1 := v1.Field(i).Interface()
		f2 := v2.Field(i).Interface()
		if !reflect.DeepEqual(f1, f2) {
			e.mismatches = append(e.mismatches,
				fmt.Sprintf("{%s: Got: %#v, Want: %#v}", t.Field(i).Name, f2, f1))
		}
	}
	return len(e.mismatches) == 0
}

func (e *equalToMatcher) String() string {
	return strings.Join(e.mismatches, ",")
}
