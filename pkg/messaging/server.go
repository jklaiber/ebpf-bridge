package messaging

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"google.golang.org/grpc"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
	"github.com/jklaiber/ebpf-bridge/pkg/logging"
	"github.com/jklaiber/ebpf-bridge/pkg/manager"
)

var log = logging.DefaultLogger.WithField("subsystem", "messaging-server")

type Server interface {
	Start() error
	Stop()
}

type MessagingServer struct {
	api.UnimplementedEbpfBridgeControllerServer
	server        *grpc.Server
	bridgeManager manager.Manager
	socket        string
	done          chan error
}

func NewMessagingServer(bridgeManager manager.Manager, socket string) *MessagingServer {
	s := &MessagingServer{
		server:        grpc.NewServer(),
		bridgeManager: bridgeManager,
		socket:        socket,
	}
	api.RegisterEbpfBridgeControllerServer(s.server, s)
	return s
}

func (s *MessagingServer) AddBridge(_ context.Context, in *api.AddCommand) (*api.AddResponse, error) {
	log.Infof("Add command received: %v", in)
	var monitor *int
	if in.Monitor != nil {
		m := int(*in.Monitor)
		monitor = &m
	}
	if err := s.bridgeManager.Add(in.Name, int(in.Iface1), int(in.Iface2), monitor); err != nil {
		log.Errorf("Failed to add bridge: %v", err)
		return &api.AddResponse{
			Success: false,
			Message: fmt.Sprintf("failed to add bridge %s: %v", in.Name, err),
		}, nil
	}
	return &api.AddResponse{
		Success: true,
		Message: fmt.Sprintf("Bridge %s added", in.Name),
	}, nil
}

func (s *MessagingServer) RemoveBridge(_ context.Context, in *api.RemoveCommand) (*api.RemoveResponse, error) {
	log.Infof("Remove command received: %v", in)
	if err := s.bridgeManager.Remove(in.Name); err != nil {
		log.Errorf("Failed to remove bridge: %v", err)
		return &api.RemoveResponse{
			Success: false,
			Message: fmt.Sprintf("failed to remove bridge %s: %v", in.Name, err),
		}, nil
	}
	return &api.RemoveResponse{
		Success: true,
		Message: fmt.Sprintf("Bridge %s removed", in.Name),
	}, nil
}

func (s *MessagingServer) ListBridges(_ context.Context, _ *api.ListCommand) (*api.ListResponse, error) {
	log.Debug("List command received")
	bridges := s.bridgeManager.List()
	descriptions := make([]*api.BridgeDescription, 0, len(bridges))
	for _, b := range bridges {
		descriptions = append(descriptions, &api.BridgeDescription{
			Name:    b.Name,
			Iface1:  b.Iface1,
			Iface2:  b.Iface2,
			Monitor: b.Monitor,
		})
	}
	return &api.ListResponse{Bridges: descriptions}, nil
}

// Start listens on the unix socket and serves requests in the background.
func (s *MessagingServer) Start() error {
	if err := os.MkdirAll(filepath.Dir(s.socket), 0o750); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}
	if err := os.Remove(s.socket); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove stale socket: %w", err)
	}
	lis, err := net.Listen("unix", s.socket)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.socket, err)
	}
	if err := os.Chmod(s.socket, socketMode); err != nil {
		_ = lis.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	s.done = make(chan error, 1)
	go func() {
		s.done <- s.server.Serve(lis)
	}()
	log.Infof("Listening on %s", s.socket)
	return nil
}

// Stop waits for running requests to finish and stops the server.
func (s *MessagingServer) Stop() {
	s.server.GracefulStop()
	if s.done != nil {
		if err := <-s.done; err != nil {
			log.Errorf("Server stopped with error: %v", err)
		}
	}
}
