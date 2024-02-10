package service

import (
	"context"

	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/logging"
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

func NewEbpfBridgeService(linkFactory hostlink.LinkFactory) *EbpfBridgeService {
	messagingServer := messaging.NewMessagingServer(linkFactory)
	return &EbpfBridgeService{
		messagingServer: messagingServer,
	}
}

func (s *EbpfBridgeService) Start() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	go func() {
		select {
		case <-s.ctx.Done():
			log.Info("Shutting down")
			s.messagingServer.Stop()
		default:
			log.Info("Starting service")
			s.messagingServer.Start()
		}
	}()
}

func (s *EbpfBridgeService) Stop() {
	s.cancel()
}
