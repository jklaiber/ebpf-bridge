// SPDX-License-Identifier: GPL-2.0

#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>

struct {
  __uint(type, BPF_MAP_TYPE_DEVMAP_HASH);
  __uint(max_entries, 3);
  __uint(key_size, sizeof(int));
  __uint(value_size, sizeof(int));
} devmap SEC(".maps");

SEC("bridge")
int xdp_bridge(struct xdp_md *ctx) {
  return bpf_redirect_map(&devmap, 0, BPF_F_EXCLUDE_INGRESS);
}

char _license[] SEC("license") = "GPL";