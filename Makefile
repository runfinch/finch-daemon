GINKGO = go run github.com/onsi/ginkgo/v2/ginkgo
# Common ginkgo options: -v for verbose mode, --focus="test name" for running single tests
GFLAGS ?= --race --randomize-all --randomize-suites
BIN = $(PWD)/bin
FINCH_ROOT ?= /Applications/Finch
FINCH_DAEMON_PROJECT_ROOT ?= $(shell pwd)

# Linux or macOS targets
.PHONY: build
build::
	$(eval PACKAGE := github.com/runfinch/finch-daemon)
	$(eval VERSION ?= $(shell git describe --match 'v[0-9]*' --dirty='.modified' --always --tags))
	$(eval GITCOMMIT := $(shell git rev-parse HEAD)$(shell if ! git diff --no-ext-diff --quiet --exit-code; then echo .m; fi))
	$(eval LDFLAGS := "-X $(PACKAGE)/version.Version=$(VERSION) -X $(PACKAGE)/version.GitCommit=$(GITCOMMIT)")
	GOOS=linux go build -ldflags $(LDFLAGS) -v -o bin/finch-daemon $(PACKAGE)/cmd/finch-daemon

# Linux targets
.PHONY: linux
linux:
ifneq ($(shell uname), Linux)
	$(error This needs to be run on linux!)
endif

.PHONY: start
start: linux build unlink
	sudo bin/finch-daemon --debug --socket-owner $${UID}

DLV=$(BIN)/dlv
$(DLV):
	GOBIN=$(BIN) go install github.com/go-delve/delve/cmd/dlv@latest

.PHONY: start-debug
start-debug: linux build $(DLV) unlink
	sudo $(DLV) --listen=:2345 --headless=true --api-version=2 exec ./bin/finch-daemon -- --debug --socket-owner $${UID}

# Unlink the unix socket if the link does not get cleaned up properly (or if finch-daemon is already running)
.PHONY: unlink
unlink: linux
ifneq ("$(wildcard /run/finch.sock)","")
	sudo unlink /run/finch.sock
endif

.PHONY: code-gen
code-gen: linux
	rm -rf ./pkg/mocks
	GOBIN=$(BIN) go install github.com/golang/mock/mockgen
	GOBIN=$(BIN) go install golang.org/x/tools/cmd/stringer
	PATH=$(BIN):$(PATH) go generate ./...
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_container/container.go -package=mocks_container github.com/containerd/containerd Container
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_container/process.go -package=mocks_container github.com/containerd/containerd Process
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_container/task.go -package=mocks_container github.com/containerd/containerd Task
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_image/store.go -package=mocks_image github.com/containerd/containerd/images Store
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_container/network_manager.go -package=mocks_container github.com/containerd/nerdctl/pkg/containerutil NetworkOptionsManager
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_cio/io.go -package=mocks_cio github.com/containerd/containerd/cio IO
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_http/response_writer.go -package=mocks_http net/http ResponseWriter
	PATH=$(BIN):$(PATH) mockgen --destination=./mocks/mocks_http/conn.go -package=mocks_http net Conn

GOLINT=$(BIN)/golangci-lint
$(GOLINT): linux
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN) v1.53.3

.PHONY: golint
golint: linux $(GOLINT) 
	$(GOLINT) run ./pkg/...

.PHONY: run-unit-tests
run-unit-tests: linux
	$(GINKGO) $(GFLAGS) ./...

# Runs tests in headless dlv mode, must specify package directory with PKG_DIR
PKG_DIR ?= .
.PHONY: debug-unit-tests
debug-unit-tests: linux $(DLV)
	sudo $(DLV) --listen=:2345 --headless=true --api-version=2 test $(PKG_DIR)

.PHONY: coverage
coverage: linux
	$(GINKGO) -r -v -race --trace --cover --coverprofile="coverage-report.out" --coverpkg=./... ./pkg/...
	go tool cover -html="coverage-report.out" -o="unit-test-coverage-report.html"

# macOS targets

.PHONY: macOS
macOS:
ifneq ($(shell uname), Darwin)
	$(error This needs to be run on macOS!)
endif

.PHONY: run-e2e-tests
run-e2e-tests: linux
	DOCKER_HOST="unix:///run/finch.sock" \
	DOCKER_API_VERSION="v1.41" \
	RUN_E2E_TESTS=1 \
	$(GINKGO) $(GFLAGS) ./e2e/...

.PHONY: release
release: linux
	@echo "$@"
	@$(FINCH_DAEMON_PROJECT_ROOT)/scripts/create-releases.sh $(RELEASE_TAG)
