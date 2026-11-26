# Copyright (c) 2024 Julian Klaiber

BINARY_NAME := ebpf-bridge
OUT_DIR     := out/bin
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS     := -s -w -X github.com/jklaiber/ebpf-bridge/ebpf-bridge/cmd.version=$(VERSION)

CLANG  ?= clang
CFLAGS := -O2 -g -Wall -Werror

# Container engine used by the container-* and lab-* targets
# (docker, nerdctl, podman). The dev image contains the full toolchain, so
# these targets also work on macOS.
CONTAINER ?= docker
DEV_IMAGE ?= ebpf-bridge-dev
LAB_NAME  ?= ebpf-bridge-lab
DEV_RUN   := $(CONTAINER) run --rm --privileged \
	-v $(CURDIR):/workspaces/ebpf-bridge -w /workspaces/ebpf-bridge \
	-v ebpf-bridge-gocache:/root/.cache -v ebpf-bridge-gomod:/go/pkg/mod \
	$(DEV_IMAGE)

# Extra flags for the e2e test binary, e.g. E2E_ARGS="-test.run TestLifecycle".
E2E_ARGS ?=
SUDO     := $(shell [ "$$(id -u)" = 0 ] || echo sudo)

.DEFAULT_GOAL := help
.PHONY: all build clean generate generate-bpf generate-go test test-coverage vuln e2e-build \
	e2e e2e-coverage lint fmt check-generated dev-image container-% lab-up lab-shell lab-down help

##@ Build

all: generate build ## Generate code and build the binary

build: ## Build the Linux binary into out/bin
	mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME) ./ebpf-bridge

clean: ## Remove build artifacts
	rm -rf out dist coverage.out e2e-coverage.out

##@ Code generation (needs clang, libbpf-dev, protoc; use container-generate on macOS)

generate: ## Regenerate BPF objects, protobuf code and mocks
	BPF_CLANG=$(CLANG) BPF_CFLAGS="$(CFLAGS)" go generate ./...

check-generated: generate ## Fail if generated files are out of date
	git diff --exit-code -- pkg/

##@ Test & lint

test: ## Run unit tests (any OS, no root needed)
	go test -race -count=1 ./...

test-coverage: ## Run unit tests with coverage
	go test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...

e2e-build: ## Build the binary and e2e test binary into out/e2e (static, any Linux)
	mkdir -p out/e2e
	CGO_ENABLED=0 GOOS=linux go build $(E2E_BUILD_FLAGS) -o out/e2e/ebpf-bridge ./ebpf-bridge
	CGO_ENABLED=0 GOOS=linux go test -c -tags e2e -o out/e2e/e2e.test ./test/e2e

e2e: e2e-build ## Run end-to-end tests with real XDP bridges (Linux; uses sudo if needed)
	$(SUDO) env EBPF_BRIDGE_BIN=$(CURDIR)/out/e2e/ebpf-bridge $(E2E_ENV) \
		out/e2e/e2e.test -test.v -test.count=1 $(E2E_ARGS)

e2e-coverage: ## Run end-to-end tests and write e2e-coverage.out
	rm -rf out/e2e/cover && mkdir -p out/e2e/cover
	$(MAKE) e2e E2E_BUILD_FLAGS="-cover -coverpkg=github.com/jklaiber/ebpf-bridge/..." \
		E2E_ENV="GOCOVERDIR=$(CURDIR)/out/e2e/cover"
	go tool covdata textfmt -i=out/e2e/cover -o e2e-coverage.out
	go tool covdata percent -i=out/e2e/cover

lint: ## Run golangci-lint
	golangci-lint run ./...

vuln: ## Check for known vulnerabilities in dependencies and the Go toolchain
	go tool govulncheck ./...

fmt: ## Format Go and C code
	golangci-lint fmt ./...
	find bpf -name '*.[ch]' | xargs clang-format -i

##@ Containers (Docker Desktop, Rancher Desktop, OrbStack, Colima, Linux)

dev-image: ## Build the development image
	$(CONTAINER) build -t $(DEV_IMAGE) .devcontainer

container-%: dev-image ## Run any target in the dev container, e.g. make container-e2e
	$(DEV_RUN) make $* VERSION=$(VERSION)

lab-up: dev-image build-in-container ## Start an interactive lab (h1, h2, mon namespaces)
	-$(CONTAINER) rm -f $(LAB_NAME) >/dev/null 2>&1
	$(CONTAINER) run -d --name $(LAB_NAME) --privileged \
		-v $(CURDIR):/workspaces/ebpf-bridge -w /workspaces/ebpf-bridge \
		$(DEV_IMAGE) test/lab/setup.sh
	@echo "Lab running. Open a shell with: make lab-shell CONTAINER=$(CONTAINER)"

lab-shell: ## Open a shell in the lab
	$(CONTAINER) exec -it $(LAB_NAME) bash

lab-down: ## Stop the lab
	-$(CONTAINER) rm -f $(LAB_NAME)

.PHONY: build-in-container
build-in-container:
	$(DEV_RUN) make build VERSION=$(VERSION)

##@ Help

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z0-9_%-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)
