FROM scratch

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/ebpf-bridge /ebpf-bridge

ENTRYPOINT ["/ebpf-bridge", "start-service"]
