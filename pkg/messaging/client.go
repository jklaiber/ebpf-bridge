package messaging

import (
	context "context"
	"time"

	grpc "google.golang.org/grpc"
)

type Client interface {
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
		conn:   conn,
		client: client,
	}
}

func connect() (*grpc.ClientConn, error) {
	// return grpc.Dial("", grpc.WithTransportCredentials(insecure.NewCredentials()))
	return grpc.Dial("unix:///tmp/ebpf_bridge.sock", grpc.WithInsecure())
}

func (mc *MessagingClient) Close() {
	if mc.conn != nil {
		if err := mc.conn.Close(); err != nil {
			panic(err)
		}
	}
}

func (mc *MessagingClient) AddBridge(in *AddCommand) (*AddResponse, error) {
	// ctx, cancel := context.WithTimeout(context.Background(), mc.timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
