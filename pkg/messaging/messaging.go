// Package messaging implements the gRPC API between the CLI and the service.
package messaging

// DefaultSocket is the unix socket the service listens on.
const DefaultSocket = "/run/ebpf-bridge.sock"

// socketMode allows root and members of the socket's group to talk to the
// service. Attaching XDP programs is a privileged operation, so the socket
// must not be world-writable.
const socketMode = 0o660
