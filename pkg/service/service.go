package service

import "github.com/jklaiber/ebpf-bridge/pkg/manager"

const SocketPath = "/tmp/ebpf-bridge.sock"

type Service interface {
	Start() error
	Stop() error
}

type EbpfBridgeService struct {
	listener      *Listener
	messages      chan string
	bridgeManager manager.Manager
}
