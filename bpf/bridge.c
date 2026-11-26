// SPDX-License-Identifier: GPL-2.0

#include <linux/types.h>
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

// Ports of one bridge: slot 0 and 1 are the bridged interfaces, slot 2 is the
// optional monitor interface. Every bridge loads its own copy of this map.
struct {
	__uint(type, BPF_MAP_TYPE_DEVMAP);
	__uint(max_entries, 3);
	__type(key, __u32);
	__type(value, __u32);
} devmap SEC(".maps");

// Forward every frame unchanged to all ports of the bridge except the one it
// arrived on. No filtering happens, so link-local frames such as LLDP, STP,
// LACP and 802.1X pass through, unlike with a regular Linux bridge.
SEC("xdp")
int xdp_bridge(struct xdp_md *ctx)
{
	return bpf_redirect_map(&devmap, 0,
							BPF_F_BROADCAST | BPF_F_EXCLUDE_INGRESS);
}

char _license[] SEC("license") = "GPL";
