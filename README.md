<h1 align="center">eBPF-Bridge</h1>
<p align="center">
    <br>
	<img src="https://img.shields.io/github/v/tag/jklaiber/ebpf-bridge.svg?label=release&logo=github&style=flat-square">
    <img src="https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat-square">
	<img src="https://img.shields.io/github/actions/workflow/status/jklaiber/ebpf-bridge/test-go.yaml?branch=main&logo=github&style=flat-square&label=tests">
	<img src="https://img.shields.io/github/actions/workflow/status/jklaiber/ebpf-bridge/lint-go.yaml?branch=main&logo=github&style=flat-square&label=checks">
    <img src="https://img.shields.io/codecov/c/github/jklaiber/ebpf-bridge
?style=flat-square&logo=codecov">
    <img src="https://img.shields.io/github/license/jklaiber/ebpf-bridge
?style=flat-square&logo=apache">
</p>

<p align="center">
</p>

---
## 🌟 Features

- **Dynamic Bridge Management:** Effortlessly add or remove network bridges to the `eBPF-Bridge` service. This flexibility allows for quick adjustments to network configurations as needs evolve.

- **Multi-Interface Bridging:** Create bridges that span multiple network interfaces, enabling complex networking setups and facilitating advanced network monitoring and manipulation tasks.

- **Live Network Monitoring:** Designate specific interfaces for monitoring by the `eBPF-Bridge` service, allowing for real-time network traffic analysis and insights.

- **Comprehensive Bridge Listing:** Easily list all bridges currently managed by the `eBPF-Bridge` service, providing administrators with a clear overview of the network's bridge topology.

- **Seamless Packet Forwarding:** Utilizing eBPF technology, `eBPF-Bridge` forwards packets without filtering, overcoming traditional limitations of Linux bridges in handling LLDP and other multicast packets essential for network discovery and management.

## 🔧 Requirements

* Linux kernel 5.0 or later
* Systemd for managing the service
* Docker for containerized deployments (optional)

## 📦 Installation

### Using Package Manager
For Debian-based systems, you can install the package using `apt`:

```
sudo apt install ./ebpf-bridge_{version}_amd64.deb
```

### Using Docker

```
docker run -d --name ebpf-bridge --privileged --network host jklaiber/ebpf-bridge:latest
```

## 🚀 Getting Started
After installation, `ebpf-bridge` is available as a system service. You can start the service using the following command:

```
sudo systemctl start ebpf-bridge.service
```
To enable the service to start on boot, use the following command:

```
sudo systemctl enable ebpf-bridge.service
```

Check the status of the service using the following command:

```
sudo systemctl status ebpf-bridge.service
```

For Docker deployments, follow the standard Docker commands to manager the container.

## ⚙️ Usage

Add a bridge to the `ebpf-bridge` service using the following command:

```
ebpf-bridge add --name test-bridge --iface1 eth0 --iface2 eth1 --monitor eth2
```

To remove a bridge from the `ebpf-bridge` service, use the following command:

```
ebpf-bridge remove --name test-bridge
```

To list all bridges managed by the `ebpf-bridge` service, use the following command:

```
ebpf-bridge list
```


## 💡 Design Rationale
The `ebpf-bridge` project addresses critical limitations in standard Linux bridges, particularly their inability to forward LLDP and certain multicast packets critical for network discovery and management. This issue stems from default bridge configurations that drop specific types of network traffic, including essential protocols like STP, LACP, and 802.1X, which are crucial for network operations. By utilizing eBPF technology, `ebpf-bridge` overcomes these limitations, as it forwards packets as they are, without any filtering or alterations. This approach ensures that all necessary network traffic, regardless of type, can be transmitted seamlessly across network segments, thus maintaining the integrity and functionality of network discovery and management protocols without the traditional constraints imposed by Linux bridge configurations.

## 🤝 Contributing
We welcome contributions to the ebpf-bridge project. If you'd like to contribute, please follow these steps:

1. Fork the repository
2. Create a new branch with your changes
3. Submit a pull request
4. Please ensure your changes are well-documented and tested.

## 📄 License
ebpf-bridge is released under the Apache-2.0 license. See [LICENSE](./LICENSE) for more information.