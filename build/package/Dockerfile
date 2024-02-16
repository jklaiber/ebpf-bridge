FROM scratch as final

COPY ebpf-bridge /ebpf-bridge

ENTRYPOINT ["/ebpf-bridge", "start-service"]
