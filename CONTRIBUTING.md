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
| `make vuln` | govulncheck: known vulnerabilities in dependencies and the Go toolchain | |
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

CI additionally runs the e2e suite in VMs with older and newer kernels (see
the `e2e-kernels` job), using [little-vm-helper](https://github.com/cilium/little-vm-helper).

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

## Dependencies and toolchain

Dependabot proposes updates for Go modules, GitHub Actions (pinned to commit
SHAs) and container base images (pinned to digests) every week. The Go
toolchain used by CI and releases is the `toolchain` line in `go.mod`; bump it
when a new Go patch release is out. `make vuln` and the weekly CI run report
known vulnerabilities in it.

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
- Pull requests are squash merged, so the **pull request title** becomes the
  commit message on `main`. It must follow
  [Conventional Commits](https://www.conventionalcommits.org), which a CI check
  enforces. The type decides whether and how the change appears in the
  changelog and how the version is bumped:

  | Title | Changelog section | Version bump |
  | --- | --- | --- |
  | `feat: add VLAN filter` | Features | minor |
  | `fix: detach XDP program on error` | Bug Fixes | patch |
  | `perf: ...` / `revert: ...` | Performance Improvements / Reverts | patch |
  | `docs: ...` | Documentation | none |
  | `refactor:`, `test:`, `build:`, `ci:`, `chore:` | not listed | none |

  Mark breaking changes with `!` (`feat!: ...`) or a `BREAKING CHANGE:` footer.
  This bumps the major version.

## Releases

Releases are automated with
[release-please](https://github.com/googleapis/release-please) and
[GoReleaser](https://goreleaser.com):

1. Every merge to `main` updates an open **release pull request** that bumps
   the version and adds the new entries to [CHANGELOG.md](./CHANGELOG.md).
2. Merging the release pull request creates the `vX.Y.Z` tag and the GitHub
   release with the changelog as release notes.
3. GoReleaser then attaches binaries, `.deb`/`.rpm` packages and SBOMs for
   amd64 and arm64 and pushes a multi-arch image to `ghcr.io/jklaiber/ebpf-bridge`.
   All artifacts get signed build provenance attestations.

To force a specific version, add a `Release-As: x.y.z` footer to a commit
message on `main`. To rebuild the artifacts of an existing tag, run the
*Release* workflow manually with that tag. The CI job "Release dry run" checks
the GoReleaser configuration on every pull request.
