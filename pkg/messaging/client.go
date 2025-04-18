//go:generate mockgen -source=client.go -destination=mocks/messaging_mock.go -package=mocks Client
package messaging

import (
	context "context"
	"fmt"
	"time"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	Close()
	AddBridge(in *api.AddCommand) (*api.AddResponse, error)
	RemoveBridge(in *api.RemoveCommand) (*api.RemoveResponse, error)
	ListBridges(in *api.ListCommand) (*api.ListResponse, error)
}

type MessagingClient struct {
	conn    *grpc.ClientConn
	client  api.EbpfBridgeControllerClient
	timeout time.Duration
}

func NewMessagingClient() *MessagingClient {
	conn, err := connect()
	if err != nil {
		panic(err)
	}

	client := api.NewEbpfBridgeControllerClient(conn)

	return &MessagingClient{
		conn:    conn,
		client:  client,
		timeout: 5 * time.Second,
	}
}

func connect() (*grpc.ClientConn, error) {
	return grpc.NewClient(
		fmt.Sprintf("%s://%s", PROTOCOL, SOCKET),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func (mc *MessagingClient) Close() {
	if mc.conn != nil {
		if err := mc.conn.Close(); err != nil {
			panic(err)
		}
	}
}

func (mc *MessagingClient) AddBridge(in *api.AddCommand) (*api.AddResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.timeout)
	defer cancel()
	return mc.client.AddBridge(ctx, in)
}

func (mc *MessagingClient) RemoveBridge(in *api.RemoveCommand) (*api.RemoveResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.timeout)
	defer cancel()
	return mc.client.RemoveBridge(ctx, in)
}

func (mc *MessagingClient) ListBridges(in *api.ListCommand) (*api.ListResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.timeout)
	defer cancel()
	return mc.client.ListBridges(ctx, in)
}
