package service

import (
	"context"

	"github.com/jklaiber/ebpf-bridge/pkg/logging"
	"github.com/jklaiber/ebpf-bridge/pkg/manager"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
)

var log = logging.DefaultLogger.WithField("subsystem", "bridge-service")

const SocketPath = "/tmp/ebpf-bridge.sock"

type Service interface {
	Start()
	Stop()
}

type EbpfBridgeService struct {
	messagingServer messaging.Server
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewEbpfBridgeService(bridgeManager manager.Manager) *EbpfBridgeService {
	messagingServer := messaging.NewMessagingServer(bridgeManager)
	return &EbpfBridgeService{
		messagingServer: messagingServer,
	}
}

func (s *EbpfBridgeService) Start() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	go func() {
		<-s.ctx.Done()
		log.Info("Shutting down")
		s.messagingServer.Stop()
	}()
	log.Info("Starting service")
	s.messagingServer.Start()
}

func (s *EbpfBridgeService) Stop() {
	s.cancel()
}
