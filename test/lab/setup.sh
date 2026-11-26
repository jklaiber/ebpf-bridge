#!/usr/bin/env bash
# Interactive lab (make lab-up): three hosts in network namespaces wired to
# veth ports in this namespace, plus a running ebpf-bridge service.
#
#   h1  [eth0 10.0.0.1] <--veth--> br-a \
#                                         ebpf-bridge
#   h2  [eth0 10.0.0.2] <--veth--> br-b /     |
#   mon [eth0]          <--veth--> br-m ------+ (monitor copy)
#
# Try:  ebpf-bridge add --name lab --iface1 br-a --iface2 br-b --monitor br-m
#       ip netns exec h1 ping 10.0.0.2
#       ip netns exec mon tcpdump -ni eth0
set -euo pipefail

ln -sf "$PWD/out/bin/ebpf-bridge" /usr/local/bin/ebpf-bridge

add_host() { # <netns> <port> [cidr]
  ip netns add "$1"
  ip link add "$2" type veth peer name eth0 netns "$1"
  ip link set "$2" up
  ip -n "$1" link set lo up
  ip -n "$1" link set eth0 up
  if [ -n "${3:-}" ]; then ip -n "$1" addr add "$3" dev eth0; fi
}

add_host h1 br-a 10.0.0.1/24
add_host h2 br-b 10.0.0.2/24
add_host mon br-m

exec ebpf-bridge start-service
