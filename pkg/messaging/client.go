package messaging

//go:generate go tool mockgen -source=client.go -destination=mocks/messaging_mock.go -package=mocks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
)

type Client interface {
	Close() error
	AddBridge(in *api.AddCommand) (*api.AddResponse, error)
	RemoveBridge(in *api.RemoveCommand) (*api.RemoveResponse, error)
	ListBridges(in *api.ListCommand) (*api.ListResponse, error)
}

type MessagingClient struct {
	conn    *grpc.ClientConn
	client  api.EbpfBridgeControllerClient
	socket  string
	timeout time.Duration
}

func NewMessagingClient(socket string) (*MessagingClient, error) {
	conn, err := grpc.NewClient(
		"unix://"+socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for %s: %w", socket, err)
	}
	return &MessagingClient{
		conn:    conn,
		client:  api.NewEbpfBridgeControllerClient(conn),
		socket:  socket,
		timeout: 5 * time.Second,
	}, nil
}

func (mc *MessagingClient) Close() error {
	if mc.conn == nil {
		return nil
	}
	return mc.conn.Close()
}

func (mc *MessagingClient) AddBridge(in *api.AddCommand) (*api.AddResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.timeout)
	defer cancel()
	resp, err := mc.client.AddBridge(ctx, in)
	return resp, mc.wrap(err)
}

func (mc *MessagingClient) RemoveBridge(in *api.RemoveCommand) (*api.RemoveResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.timeout)
	defer cancel()
	resp, err := mc.client.RemoveBridge(ctx, in)
	return resp, mc.wrap(err)
}

func (mc *MessagingClient) ListBridges(in *api.ListCommand) (*api.ListResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.timeout)
	defer cancel()
	resp, err := mc.client.ListBridges(ctx, in)
	return resp, mc.wrap(err)
}

// ErrServiceUnavailable is returned when the service cannot be reached.
var ErrServiceUnavailable = errors.New("ebpf-bridge service is not reachable")

func (mc *MessagingClient) wrap(err error) error {
	if err == nil {
		return nil
	}
	if status.Code(err) == codes.Unavailable {
		return fmt.Errorf("%w at %s (is it running?): %v", ErrServiceUnavailable, mc.socket, status.Convert(err).Message())
	}
	return err
}
