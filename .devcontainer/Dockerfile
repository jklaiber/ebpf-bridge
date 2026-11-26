# Development and test image for ebpf-bridge.
#
# Used by the VS Code dev container and by the `make container-*` targets, so
# the whole toolchain (Go, clang/libbpf for the BPF program, protoc,
# golangci-lint) works the same on Linux and macOS (amd64 and arm64).
ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}-trixie

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       clang clang-format llvm libbpf-dev linux-libc-dev \
       protobuf-compiler make git sudo \
       iproute2 iputils-ping tcpdump lldpd procps \
    && rm -rf /var/lib/apt/lists/* \
    # BPF programs include <asm/types.h>, which lives in the multiarch dir.
    && ln -sf "/usr/include/$(gcc -dumpmachine)/asm" /usr/include/asm

# bpf2go, mockgen and the protoc plugins are pinned in go.mod (`go tool`).
# Keep in sync with .github/workflows/ci.yaml.
ARG GOLANGCI_LINT_VERSION=v2.15.0
RUN GOBIN=/usr/local/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} \
    && rm -rf /root/.cache /go/pkg/mod

ENV GOFLAGS=-buildvcs=false
WORKDIR /workspaces/ebpf-bridge
