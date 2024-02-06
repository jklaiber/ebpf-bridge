# Copyright (c) 2024 Julian Klaiber

GOCMD=go
BINARY_NAME=ebpf-bridge
CLANG ?= clang
CFLAGS :=  -O2 -g -Wall $(CFLAGS) -DDEBUG

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all build clean

all: go-gen build ## Build the entire project

install-deps: ## Install development dependencies
	go install honnef.co/go/tools/cmd/staticcheck@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
	sudo apt install clang clang-format llvm gcc libbpf-dev libelf-dev make linux-headers-$(uname -r)
	sudo ln -s /usr/include/x86_64-linux-gnu/asm /usr/include/asm
# https://github.com/xdp-project/xdp-tools
# https://github.com/libbpf/bpftool/blob/master/README.md

build: ## Compile the Go binary
	mkdir -p out/bin
	$(GOCMD) build -o out/bin/$(BINARY_NAME) ./$(BINARY_NAME)/main.go

clean: ## Clean build artifacts
	rm -fr out

go-gen: export BPF_CLANG := $(CLANG)
go-gen: export BPF_CFLAGS := $(CFLAGS)
go-gen: ## Generate BPF code and Go bindings
	go generate ./...

test: ## Run go tests
	go clean -testcache
	go test ./...

test-coverage: ## Run go tests with coverage
	go clean -testcache
	go test ./... -coverprofile=coverage.out

fix-clang-style: ## Fix the clang style
	find . -iname *.h -o -iname *.c | xargs clang-format -i

proto-gen: ## Generate protobuf files
	protoc --go_out=. --go_opt=Mpkg/messaging/messaging.proto=pkg/messaging --go-grpc_out=. --go-grpc_opt=Mpkg/messaging/messaging.proto=pkg/messaging pkg/messaging/messaging.proto --experimental_allow_proto3_optional
	
help: ## Show this help message
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_0-9-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)