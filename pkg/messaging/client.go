//go:generate mockgen -source=client.go -destination=mocks/messaging_mock.go -package=mocks Client
package messaging

import (
	context "context"
	"fmt"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	Close()
	AddBridge(in *AddCommand) (*AddResponse, error)
	RemoveBridge(in *RemoveCommand) (*RemoveResponse, error)
	ListBridges(in *ListCommand) (*ListResponse, error)
}

type MessagingClient struct {
	conn    *grpc.ClientConn
	client  EbpfBridgeControllerClient
	timeout time.Duration
}

func NewMessagingClient() *MessagingClient {
	conn, err := connect()
	if err != nil {
		panic(err)
	}

	client := NewEbpfBridgeControllerClient(conn)

	return &MessagingClient{
		conn:    conn,
		client:  client,
		timeout: 5 * time.Second,
	}
}

func connect() (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf("%s://%s", PROTOCOL, SOCKET), grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func (mc *MessagingClient) Close() {
	if mc.conn != nil {
		if err := mc.conn.Close(); err != nil {
			panic(err)
		}
	}
}

func (mc *MessagingClient) AddBridge(in *AddCommand) (*AddResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.timeout)
	defer cancel()
	return mc.client.AddBridge(ctx, in)
}

func (mc *MessagingClient) RemoveBridge(in *RemoveCommand) (*RemoveResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.timeout)
	defer cancel()
	return mc.client.RemoveBridge(ctx, in)
}

func (mc *MessagingClient) ListBridges(in *ListCommand) (*ListResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.timeout)
	defer cancel()
	return mc.client.ListBridges(ctx, in)
}
