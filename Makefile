# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

GINKGO = go run github.com/onsi/ginkgo/v2/ginkgo
# Common ginkgo options: -v for verbose mode, --focus="test name" for running single tests
GFLAGS ?= --race --randomize-all --randomize-suites 
BIN = $(PWD)/bin
FINCH_DAEMON_PROJECT_ROOT ?= $(shell pwd)

# Base path used to install.
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin

CRED_HELPER_PREFIX ?= /usr
CRED_HELPER_BINDIR ?= $(CRED_HELPER_PREFIX)/bin

BINARY = $(addprefix bin/,finch-daemon)
CREDENTIAL_HELPER = $(addprefix bin/,docker-credential-finch)

PACKAGE := github.com/runfinch/finch-daemon
VERSION ?= $(shell git describe --match 'v[0-9]*' --dirty='.modified' --always --tags)
GITCOMMIT := $(shell git rev-parse HEAD)$(shell if ! git diff --no-ext-diff --quiet --exit-code; then echo .m; fi)

ifndef GODEBUG
	EXTRA_LDFLAGS += -s -w
endif

LDFLAGS_BASE := -X $(PACKAGE)/version.Version=$(VERSION) -X $(PACKAGE)/version.GitCommit=$(GITCOMMIT) $(EXTRA_LDFLAGS)

.PHONY: build
build: build-daemon build-credential-helper

.PHONY: build-daemon
build-daemon:
ifeq ($(STATIC),)
	@echo "Building Dynamic Binary"
	CGO_ENABLED=1 GOOS=linux go build \
		-ldflags "$(LDFLAGS_BASE)" \
		-v -o $(BINARY) $(PACKAGE)/cmd/finch-daemon
else
	@echo "Building Static Binary"
	CGO_ENABLED=0 GOOS=linux go build \
		-tags netgo \
		-ldflags "$(LDFLAGS_BASE) -extldflags '-static'" \
		-v -o $(BINARY) $(PACKAGE)/cmd/finch-daemon
endif

.PHONY: build-credential-helper
build-credential-helper:
ifeq ($(STATIC),)
	@echo "Building Dynamic Credential Helper"
	CGO_ENABLED=1 GOOS=linux go build \
		-ldflags "$(LDFLAGS_BASE)" \
		-v -o $(CREDENTIAL_HELPER) $(PACKAGE)/cmd/finch-credential-helper
else
	@echo "Building Static Credential Helper"
	CGO_ENABLED=0 GOOS=linux go build \
		-tags netgo \
		-ldflags "$(LDFLAGS_BASE) -extldflags '-static'" \
		-v -o $(CREDENTIAL_HELPER) $(PACKAGE)/cmd/finch-credential-helper
endif

clean:
	@rm -f $(BINARIES)
	@rm -rf $(BIN)

.PHONY: linux
linux:
ifneq ($(shell uname), Linux)
	$(error This needs to be run on linux!)
endif

.PHONY: start
start: linux build unlink
	sudo $(BINARY) --debug --socket-owner $${UID}

DLV=$(BIN)/dlv
$(DLV):
	GOBIN=$(BIN) go install github.com/go-delve/delve/cmd/dlv@latest

.PHONY: start-debug
start-debug: linux build $(DLV) unlink
	sudo $(DLV) --listen=:2345 --headless=true --api-version=2 exec $(BINARY) -- --debug --socket-owner $${UID}

install: linux
	install -d $(DESTDIR)$(BINDIR)
	install $(BINARY) $(DESTDIR)$(BINDIR)
	install $(CREDENTIAL_HELPER) $(DESTDIR)$(CRED_HELPER_BINDIR)

uninstall:
	@rm -f $(addprefix $(DESTDIR)$(BINDIR)/,$(notdir $(BINARY)))
	@rm -f $(addprefix $(DESTDIR)$(CRED_HELPER_BINDIR)/,$(notdir $(CREDENTIAL_HELPER)))

# Unlink the unix socket if the link does not get cleaned up properly (or if finch-daemon is already running)
.PHONY: unlink
unlink: linux
ifneq ("$(wildcard /run/finch.sock)","")
	sudo unlink /run/finch.sock
endif
ifneq ("$(wildcard /run/finch/credential.sock)","")
	sudo unlink /run/finch/credential.sock
endif

.PHONY:  gen-code
 gen-code: linux
	rm -rf ./pkg/mocks
	GOBIN=$(BIN) go install go.uber.org/mock/mockgen@v0.5.2
	GOBIN=$(BIN) go install golang.org/x/tools/cmd/stringer@v0.31.0
	PATH=$(BIN):$(PATH) go generate ./...
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_container/container.go -package=mocks_container github.com/containerd/containerd/v2/client Container
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_container/process.go -package=mocks_container github.com/containerd/containerd/v2/client Process
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_container/task.go -package=mocks_container github.com/containerd/containerd/v2/client Task
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_image/store.go -package=mocks_image github.com/containerd/containerd/v2/core/images Store
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_container/network_manager.go -package=mocks_container github.com/containerd/nerdctl/v2/pkg/containerutil NetworkOptionsManager
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_cio/io.go -package=mocks_cio github.com/containerd/containerd/v2/pkg/cio IO
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_http/response_writer.go -package=mocks_http net/http ResponseWriter
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_http/conn.go -package=mocks_http net Conn

GOLINT=$(BIN)/golangci-lint
$(GOLINT): linux
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN) v1.60.3

.PHONY: lint
lint: linux $(GOLINT)
	$(GOLINT) run ./...

.PHONY: test-unit
test-unit: linux
	$(GINKGO) $(GFLAGS) ./...

# Runs tests in headless dlv mode, must specify package directory with PKG_DIR
PKG_DIR ?= .
.PHONY: test-unit-debug
test-unit-debug: linux $(DLV)
	sudo $(DLV) --listen=:2345 --headless=true --api-version=2 test $(PKG_DIR)

.PHONY: test-e2e
test-e2e: linux
	DOCKER_HOST="unix:///run/finch.sock" \
	DOCKER_API_VERSION="v1.41" \
	TEST_E2E=1 \
	$(GINKGO) $(GFLAGS) ./e2e/...

.PHONY: test-e2e-opa
test-e2e-opa: linux
	DOCKER_HOST="unix:///run/finch.sock" \
	DOCKER_API_VERSION="v1.41" \
	MIDDLEWARE_E2E=1 \
	TEST_E2E=0 \
	DAEMON_ROOT="$(BIN)/finch-daemon" \
	$(GINKGO) $(GFLAGS) ./e2e/...

.PHONY: licenses
licenses:
	PATH=$(BIN):$(PATH) go-licenses report --template="scripts/third-party-license.tpl" --ignore github.com/runfinch ./... > THIRD_PARTY_LICENSES

.PHONY: coverage
coverage: linux
	$(GINKGO) -r -v -race --trace --cover --coverprofile="coverage-report.out" ./...
	go tool cover -html="coverage-report.out" -o="unit-test-coverage-report.html"

.PHONY: release
release: linux
	@echo "$@"
	@$(FINCH_DAEMON_PROJECT_ROOT)/scripts/create-releases.sh $(RELEASE_TAG)

.PHONY: macos
macos:
ifeq ($(shell uname), Darwin)
	@echo "Running on macOS"
else
	$(error This target can only be run on macOS!)
endif
	

DAEMON_DOCKER_HOST := "unix:///Applications/Finch/lima/data/finch/sock/finch.sock"
# DAEMON_ROOT

.PHONY: test-e2e-inside-vm
test-e2e-inside-vm: macos
	DOCKER_HOST=$(DAEMON_DOCKER_HOST) \
	DOCKER_API_VERSION="v1.41" \
	TEST_E2E=1 \
	go test ./e2e -test.v -ginkgo.v -ginkgo.randomize-all \
	--subject="finch" \
	--daemon-context-subject-prefix="/Applications/Finch/lima/bin/limactl shell finch sudo" \
	--daemon-context-subject-env="LIMA_HOME=/Applications/Finch/lima/data"
