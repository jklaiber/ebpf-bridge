FROM --platform=linux/amd64 mcr.microsoft.com/vscode/devcontainers/base:0-jammy

ARG DEBIAN_FRONTEND=noninteractive

# Install core dev tools + Linux headers (for eBPF, VPP, etc.)
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    git \
    vim \
    ca-certificates \
    locales 

# Generate and set UTF‑8 locale
RUN locale-gen en_US.UTF-8
ENV LANG=en_US.UTF-8 LC_ALL=en_US.UTF-8

# Install Go (adjust GO_VERSION as needed)
ARG GO_VERSION=1.24.2
RUN curl -fsSL "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" \
    | tar -C /usr/local -xz \
    && mv /usr/local/go /usr/local/go-${GO_VERSION} \
    && ln -s /usr/local/go-${GO_VERSION} /usr/local/go

ENV GOROOT=/usr/local/go
ENV GOPATH=/workspaces/go
ENV PATH=$GOROOT/bin:$GOPATH/bin:$PATH

# Install golangci-lint
ARG GOLANGCI_LINT_VERSION=2.1.2
RUN curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" \
    | sh -s -- -b /usr/local/bin v${GOLANGCI_LINT_VERSION} \
    && golangci-lint version

# Prepare workspaces
RUN mkdir -p /workspaces && chown vscode:vscode /workspaces
WORKDIR /workspaces

# Switch back to the 'vscode' user
USER vscode
