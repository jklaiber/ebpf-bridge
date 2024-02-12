package messaging

import (
	context "context"
	"fmt"
	"net"
	"os"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
	"github.com/jklaiber/ebpf-bridge/pkg/bridge"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/logging"
	"github.com/jklaiber/ebpf-bridge/pkg/manager"
	grpc "google.golang.org/grpc"
)

var log = logging.DefaultLogger.WithField("subsystem", "messaging-server")

type Server interface {
	Start()
	Stop()
	AddBridge(ctx context.Context, in *api.AddCommand) (*api.AddResponse, error)
	RemoveBridge(ctx context.Context, in *api.RemoveCommand) (*api.RemoveResponse, error)
	ListBridges(ctx context.Context, in *api.ListCommand) (*api.ListResponse, error)
}

type MessagingServer struct {
	api.UnimplementedEbpfBridgeControllerServer
	server        *grpc.Server
	bridgeManager manager.Manager
}

func NewMessagingServer(linkFactory hostlink.LinkFactory, bridgeFactory bridge.BridgeFactory) *MessagingServer {
	return &MessagingServer{
		server:        grpc.NewServer(),
		bridgeManager: manager.NewBridgeManager(linkFactory, bridgeFactory),
	}
}

func (s *MessagingServer) AddBridge(ctx context.Context, in *api.AddCommand) (*api.AddResponse, error) {
	log.Infof("Add command received: %v", in)
	if in.Monitor != nil {
		log.Info("Monitor is not nil")
		monitorValue := int(*in.Monitor)
		err := s.bridgeManager.Add(in.Name, int(in.Iface1), int(in.Iface2), &monitorValue)
		if err != nil {
			log.Errorf("Failed to add bridge: %v", err)
			return &api.AddResponse{
				Success: false,
				Message: fmt.Sprintf("failed to add bridge: %v", err),
			}, nil
		}
	} else {
		err := s.bridgeManager.Add(in.Name, int(in.Iface1), int(in.Iface2), nil)
		if err != nil {
			log.Errorf("Failed to add bridge: %v", err)
			return &api.AddResponse{
				Success: false,
				Message: fmt.Sprintf("failed to add bridge: %v", err),
			}, nil
		}
	}
	return &api.AddResponse{
		Success: true,
		Message: fmt.Sprintf("Bridge %s added", in.Name),
	}, nil
}

func (s *MessagingServer) RemoveBridge(ctx context.Context, in *api.RemoveCommand) (*api.RemoveResponse, error) {
	log.Infof("Remove command received: %v", in)
	err := s.bridgeManager.Remove(in.Name)
	if err != nil {
		log.Errorf("Failed to remove bridge: %v", err)
		return &api.RemoveResponse{
			Success: false,
			Message: fmt.Sprintf("failed to remove bridge: %v", err),
		}, nil
	}
	return &api.RemoveResponse{
		Success: true,
		Message: fmt.Sprintf("Bridge %s removed", in.Name),
	}, nil
}

func (s *MessagingServer) ListBridges(ctx context.Context, in *api.ListCommand) (*api.ListResponse, error) {
	log.Println("List command received")
	bridges := s.bridgeManager.List()
	var bridgeDescriptions []*api.BridgeDescription
	for _, bridge := range bridges {
		bridgeDescriptions = append(bridgeDescriptions, &api.BridgeDescription{
			Name:    bridge.Name,
			Iface1:  bridge.Iface1,
			Iface2:  bridge.Iface2,
			Monitor: bridge.Monitor,
		})
	}
	return &api.ListResponse{Bridges: bridgeDescriptions}, nil
}

func (s *MessagingServer) Start() {
	os.Remove(SOCKET)

	api.RegisterEbpfBridgeControllerServer(s.server, s)
	lis, err := net.Listen(PROTOCOL, SOCKET)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := os.Chmod(SOCKET, 0777); err != nil {
		log.Fatalf("failed to change socket permissions: %v", err)
	}

	go func() {
		if err := s.server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

func (s *MessagingServer) Stop() {
	s.server.GracefulStop()
}
