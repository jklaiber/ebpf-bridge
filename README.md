<h1 align="center">eBPF-Bridge</h1>
<p align="center">
    <a href="https://github.com/jklaiber/ebpf-bridge/releases"><img src="https://img.shields.io/github/v/tag/jklaiber/ebpf-bridge.svg?label=release&logo=github&style=flat-square"></a>
    <a href="https://github.com/jklaiber/ebpf-bridge/actions/workflows/ci.yaml"><img src="https://img.shields.io/github/actions/workflow/status/jklaiber/ebpf-bridge/ci.yaml?branch=main&logo=github&style=flat-square&label=ci"></a>
    <a href="https://codecov.io/gh/jklaiber/ebpf-bridge"><img src="https://img.shields.io/codecov/c/github/jklaiber/ebpf-bridge?style=flat-square&logo=codecov"></a>
    <a href="https://scorecard.dev/viewer/?uri=github.com/jklaiber/ebpf-bridge"><img src="https://img.shields.io/ossf-scorecard/github.com/jklaiber/ebpf-bridge?style=flat-square&label=openssf%20scorecard"></a>
    <a href="https://goreportcard.com/report/github.com/jklaiber/ebpf-bridge"><img src="https://goreportcard.com/badge/github.com/jklaiber/ebpf-bridge?style=flat-square"></a>
    <a href="./LICENSE"><img src="https://img.shields.io/github/license/jklaiber/ebpf-bridge?style=flat-square"></a>
</p>

<p align="center">
Transparent Layer 2 bridges between network interfaces, implemented with XDP.<br>
Unlike a Linux bridge, <code>ebpf-bridge</code> forwards <em>every</em> frame, including LLDP, STP, LACP and 802.1X.
</p>

---
## 🌟 Features

- **Dynamic Bridge Management:** Effortlessly add or remove network bridges to the `eBPF-Bridge` service. This flexibility allows for quick adjustments to network configurations as needs evolve.

- **Multiple Independent Bridges:** Run any number of isolated bridges side by side, each connecting two network interfaces.

- **Live Network Monitoring:** Designate specific interfaces for monitoring by the `eBPF-Bridge` service, allowing for real-time network traffic analysis and insights.

- **Comprehensive Bridge Listing:** Easily list all bridges currently managed by the `eBPF-Bridge` service, providing administrators with a clear overview of the network's bridge topology.

- **Seamless Packet Forwarding:** Utilizing eBPF technology, `eBPF-Bridge` forwards packets without filtering, overcoming traditional limitations of Linux bridges in handling LLDP and other multicast packets essential for network discovery and management.

## 🔧 Requirements

* Linux kernel 5.13 or later (XDP devmap broadcast, `BPF_F_EXCLUDE_INGRESS`).
  CI runs the end-to-end tests on kernels 5.15, 6.1, 6.6, 6.12 and 6.18.
* amd64 or arm64
* root privileges for the service
* systemd (for the packaged service) or Docker

## 📦 Installation

### Packages

Download the `.deb` or `.rpm` for your architecture from the
[releases page](https://github.com/jklaiber/ebpf-bridge/releases) and install it.
See the [changelog](./CHANGELOG.md) for what changed in each release.

```
sudo apt install ./ebpf-bridge_<version>_linux_amd64.deb
# or
sudo dnf install ./ebpf-bridge_<version>_linux_amd64.rpm
```

The package installs and starts the `ebpf-bridge` systemd service:

```
sudo systemctl status ebpf-bridge.service
```

### Docker

```
docker run -d --name ebpf-bridge --privileged --network host ghcr.io/jklaiber/ebpf-bridge:2
```

Images are tagged `latest`, `<major>`, `<major>.<minor>` and `<version>`.
Run the CLI inside the container:

```
docker exec ebpf-bridge /ebpf-bridge list
```

### Verifying a release

Every release archive, package and container image has a signed
[build provenance attestation](https://docs.github.com/actions/security-for-github-actions/using-artifact-attestations)
and an SBOM. Verify a download or the image with the GitHub CLI:

```
gh attestation verify ebpf-bridge_<version>_linux_amd64.deb --repo jklaiber/ebpf-bridge
gh attestation verify oci://ghcr.io/jklaiber/ebpf-bridge:<version> --repo jklaiber/ebpf-bridge
```

### From source

```
make build   # writes out/bin/ebpf-bridge
```

## ⚙️ Usage

Bridge `eth0` and `eth1` and mirror all bridged traffic to `eth2`:

```
sudo ebpf-bridge add --name test-bridge --iface1 eth0 --iface2 eth1 --monitor eth2
```

The monitor interface is optional and receive-only: frames sent on it are not
forwarded into the bridge.

List and remove bridges:

```
sudo ebpf-bridge list
sudo ebpf-bridge remove --name test-bridge
```

The CLI talks to the service over `/run/ebpf-bridge.sock`. Use `--socket` to
choose a different path, and `ebpf-bridge start-service --log-level debug` for
verbose service logs. When the service stops, all bridges are removed and the
XDP programs are detached.

## 🧭 How it works

For every bridge, the service loads its own instance of a small XDP program
([bpf/bridge.c](bpf/bridge.c)) with a device map holding the bridge's ports. The
program is attached to both bridged interfaces in generic mode, so it works with
every driver, including veth. Each received frame is redirected unchanged to all
ports of the bridge except the one it arrived on.

## 💡 Design Rationale
The `ebpf-bridge` project addresses critical limitations in standard Linux bridges, particularly their inability to forward LLDP and certain multicast packets critical for network discovery and management. This issue stems from default bridge configurations that drop specific types of network traffic, including essential protocols like STP, LACP, and 802.1X, which are crucial for network operations. By utilizing eBPF technology, `ebpf-bridge` overcomes these limitations, as it forwards packets as they are, without any filtering or alterations. This approach ensures that all necessary network traffic, regardless of type, can be transmitted seamlessly across network segments, thus maintaining the integrity and functionality of network discovery and management protocols without the traditional constraints imposed by Linux bridge configurations.

## 🤝 Contributing

Contributions are welcome! [CONTRIBUTING.md](./CONTRIBUTING.md) explains how to
build, test (also on macOS, using containers) and submit changes. Please report
security issues as described in [SECURITY.md](./SECURITY.md).

## 📄 License
ebpf-bridge is released under the Apache-2.0 license. See [LICENSE](./LICENSE) for more information.