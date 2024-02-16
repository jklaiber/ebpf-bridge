FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN 

FROM scratch as final

COPY --from=builder /app/ebpf-bridge /ebpf-bridge

ENTRYPOINT ["/ebpf-bridge", "start-service"]
