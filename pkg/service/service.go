// Package service wires the bridge manager to the API server.
package service

import (
	"github.com/jklaiber/ebpf-bridge/pkg/logging"
	"github.com/jklaiber/ebpf-bridge/pkg/manager"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
)

var log = logging.DefaultLogger.WithField("subsystem", "bridge-service")

type EbpfBridgeService struct {
	server  messaging.Server
	manager manager.Manager
}

func NewEbpfBridgeService(bridgeManager manager.Manager, server messaging.Server) *EbpfBridgeService {
	return &EbpfBridgeService{
		server:  server,
		manager: bridgeManager,
	}
}

func (s *EbpfBridgeService) Start() error {
	log.Info("Starting service")
	return s.server.Start()
}

// Stop stops accepting requests and removes all bridges, so no XDP program
// is left attached after the service exits.
func (s *EbpfBridgeService) Stop() error {
	log.Info("Shutting down")
	s.server.Stop()
	return s.manager.RemoveAll()
}
