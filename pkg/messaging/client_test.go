package messaging

import (
	"testing"
	"time"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
	"github.com/jklaiber/ebpf-bridge/pkg/api/mocks"
	"go.uber.org/mock/gomock"
)

func TestMessagingClient_AddBridge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ebpfBridgeControllerMock := mocks.NewMockEbpfBridgeControllerClient(ctrl)
	mc := &MessagingClient{
		conn:    nil,
		client:  ebpfBridgeControllerMock,
		timeout: 1 * time.Second,
	}

	in := &api.AddCommand{
		Name:   "test",
		Iface1: 1,
		Iface2: 2,
	}

	ebpfBridgeControllerMock.EXPECT().AddBridge(gomock.Any(), in).Return(&api.AddResponse{}, nil)

	if _, err := mc.AddBridge(in); err != nil {
		t.Errorf("AddBridge() error = %v", err)
	}
}

func TestMessagingClient_RemoveBridge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ebpfBridgeControllerMock := mocks.NewMockEbpfBridgeControllerClient(ctrl)
	mc := &MessagingClient{
		conn:    nil,
		client:  ebpfBridgeControllerMock,
		timeout: 1 * time.Second,
	}

	in := &api.RemoveCommand{
		Name: "test",
	}

	ebpfBridgeControllerMock.EXPECT().RemoveBridge(gomock.Any(), in).Return(&api.RemoveResponse{}, nil)

	if _, err := mc.RemoveBridge(in); err != nil {
		t.Errorf("RemoveBridge() error = %v", err)
	}
}

func TestMessagingClient_ListBridges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ebpfBridgeControllerMock := mocks.NewMockEbpfBridgeControllerClient(ctrl)
	mc := &MessagingClient{
		conn:    nil,
		client:  ebpfBridgeControllerMock,
		timeout: 1 * time.Second,
	}

	in := &api.ListCommand{}

	ebpfBridgeControllerMock.EXPECT().ListBridges(gomock.Any(), in).Return(&api.ListResponse{}, nil)

	if _, err := mc.ListBridges(in); err != nil {
		t.Errorf("ListBridges() error = %v", err)
	}
}
