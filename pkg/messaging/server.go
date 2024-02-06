package messaging

import (
	context "context"
	"net"
	"os"

	"github.com/jklaiber/ebpf-bridge/pkg/bridge"
	"github.com/jklaiber/ebpf-bridge/pkg/logging"
	grpc "google.golang.org/grpc"
)

var log = logging.DefaultLogger.WithField("subsystem", "messaging-server")

type Server interface {
	Start()
	Stop()
	AddBridge(ctx context.Context, in *AddCommand) (*AddResponse, error)
	RemoveBridge(ctx context.Context, in *RemoveCommand) (*RemoveResponse, error)
	ListBridges(ctx context.Context, in *ListCommand) (*ListResponse, error)
}

type MessagingServer struct {
	UnimplementedEbpfBridgeControllerServer
	server        *grpc.Server
	bridgeManager bridge.Manager
}

func NewMessagingServer() *MessagingServer {
	return &MessagingServer{
		server:        grpc.NewServer(),
		bridgeManager: bridge.NewBridgeManager(),
	}
}

func (s *MessagingServer) AddBridge(ctx context.Context, in *AddCommand) (*AddResponse, error) {
	log.Infof("Add command received: %v", in)
	err := s.bridgeManager.Add(in.Name, int(in.Iface1), int(in.Iface2), int(5))
	if err != nil {
		return &AddResponse{Success: false}, nil
	}
	return &AddResponse{Success: true}, nil
}

func (s *MessagingServer) RemoveBridge(ctx context.Context, in *RemoveCommand) (*RemoveResponse, error) {
	log.Infof("Remove command received: %v", in)
	err := s.bridgeManager.Remove(in.Name)
	if err != nil {
		return &RemoveResponse{Success: false}, nil
	}
	return &RemoveResponse{Success: true}, nil
}

func (s *MessagingServer) ListBridges(ctx context.Context, in *ListCommand) (*ListResponse, error) {
	log.Printf("List command received: %v", in)
	return &ListResponse{Bridges: []*BridgeDescription{}}, nil
}

func (s *MessagingServer) Start() {
	os.Remove(SOCKET)
	RegisterEbpfBridgeControllerServer(s.server, s)
	lis, err := net.Listen(PROTOCOL, SOCKET)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
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
