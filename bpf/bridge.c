// SPDX-License-Identifier: GPL-2.0

#include <linux/types.h>
#include <linux/bpf.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

struct {
  __uint(type, BPF_MAP_TYPE_DEVMAP);
  __uint(max_entries, 3);
  __type(key, __u32);
  __type(value, __u32);
  __uint(pinning, LIBBPF_PIN_BY_NAME);
} devmap SEC(".maps");

SEC("xdp")
int xdp_bridge(struct xdp_md *ctx) {
  return bpf_redirect_map(&devmap, 0, BPF_F_BROADCAST | BPF_F_EXCLUDE_INGRESS);
}

char _license[] SEC("license") = "GPL";