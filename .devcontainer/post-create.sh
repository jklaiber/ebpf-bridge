#!/usr/bin/env bash
set -euo pipefail

sudo apt-get install -y --no-install-recommends protobuf-compiler

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go mod tidy

echo "✔ protoc version: $(protoc --version)"
echo "✔ protoc-gen-go:   $(which protoc-gen-go)"
echo "✔ protoc-gen-go-grpc: $(which protoc-gen-go-grpc)"
