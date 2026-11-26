# Contributing to ebpf-bridge

Thanks for helping! Bug reports, ideas and pull requests are welcome.

- **Bugs and feature requests:** open an [issue](https://github.com/jklaiber/ebpf-bridge/issues).
- **Security issues:** do not open an issue, follow [SECURITY.md](./SECURITY.md).
- **Changes:** open a pull request. Every change, including the maintainer's,
  goes through a pull request and needs a green CI.

## Development environment

You need [Go](https://go.dev/dl/) (see `go.mod` for the version) and `make`.
Run `make help` for all targets.

Everything else (clang and libbpf for the BPF program, protoc, golangci-lint,
iproute2 for the e2e tests) is part of the development image in
[.devcontainer](./.devcontainer). You can use it in three ways:

- **VS Code / dev containers:** "Reopen in Container".
- **Any container engine:** prefix a target with `container-`, e.g.
  `make container-e2e`. Set `CONTAINER=nerdctl` or `CONTAINER=podman` if you do
  not use `docker`.
- **Natively on Linux:** install the tools listed in the Dockerfile.

### macOS

The service needs a Linux kernel, but Docker Desktop, Rancher Desktop, OrbStack
and Colima all run containers in a Linux VM whose kernel supports XDP. The
privileged dev container uses that kernel, so the full test suite runs on a Mac:

```
make test              # unit tests, natively on macOS
make container-e2e     # end-to-end tests in the Linux VM
make container-lint
```

## Tests

| Command | What it tests | Needs |
| --- | --- | --- |
| `make test` | Unit tests of all packages with mocks (`go test -race ./...`) | Any OS |
| `make e2e` | The real binary: builds it, starts services, creates network namespaces and sends frames through XDP bridges | Linux, root (uses `sudo`) |
| `make container-e2e` | Same, in the dev container | Docker or similar |
| `make lint` | golangci-lint (config in `.golangci.yml`) | |
| `make check-generated` | Generated code matches its sources | clang, protoc |

The end-to-end tests ([test/e2e](./test/e2e)) build a topology of three hosts
(`h1`, `h2`, `mon`) per test, each in its own network namespace, and check:

- forwarding of link-local control frames (LLDP, STP, LACP, 802.1X, CDP),
  broadcast, multicast, unicast, VLAN-tagged and full-MTU frames in both
  directions, and that frames are not reflected to the sender
- mirroring to the monitor port, which is receive-only
- IP connectivity, isolation between multiple bridges
- the CLI: add, list, remove and error handling
- that no XDP program stays attached when the service stops or is killed

Run a single test with `make e2e E2E_ARGS="-test.run TestLifecycle"`.
`make e2e-coverage` also records coverage of the service and CLI processes.

### Interactive lab

To try things by hand, start a lab with the same topology and a running service:

```
make lab-up
make lab-shell
# inside the lab:
ebpf-bridge add --name lab --iface1 br-a --iface2 br-b --monitor br-m
ip netns exec h1 ping 10.0.0.2
ip netns exec mon tcpdump -ni eth0
```

Stop it with `make lab-down`.

## Code generation

Generated files are committed, so a normal build only needs Go. After changing
one of the sources below, run `make generate` (or `make container-generate`) and
commit the result. CI fails if generated files are out of date.

| Source | Generated |
| --- | --- |
| `bpf/bridge.c` | `pkg/bpf/bpf_bpfe[lb].{go,o}` (bpf2go) |
| `pkg/api/messaging.proto` | `pkg/api/*.pb.go` |
| interfaces with a `//go:generate go tool mockgen` line | `*/mocks/*.go` |

The code generators are pinned as Go tools in `go.mod`.

## Pull requests

- Keep changes focused and add tests for new behavior and bug fixes.
- Make sure `make test`, `make lint` and, for changes to the service, CLI or BPF
  program, `make e2e` pass.
- Use clear commit messages in the imperative mood ("Add ...", "Fix ...").

## Releases

Pushing a tag `vX.Y.Z` runs [GoReleaser](./.goreleaser.yaml), which publishes
binaries, `.deb`/`.rpm` packages for amd64 and arm64 and a multi-arch image to
`ghcr.io/jklaiber/ebpf-bridge`. The CI job "Release dry run" checks this on
every pull request.
