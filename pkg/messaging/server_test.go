package messaging

import (
	"context"
	"fmt"
	"testing"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
	"github.com/jklaiber/ebpf-bridge/pkg/manager/mocks"
	"go.uber.org/mock/gomock"
)

func TestMessagingServer_New(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bridgeManager := mocks.NewMockManager(ctrl)
	server := NewMessagingServer(bridgeManager)
	if server == nil {
		t.Error("NewMessagingServer() returned nil")
	}
}

func TestMessagingServer_AddBridge(t *testing.T) {
	tests := []struct {
		name     string
		mockFunc func(*mocks.MockManager)
		testCall func(*MessagingServer) (*api.AddResponse, error)
		wantErr  bool
	}{
		{
			name: "AddBridge() without monitorIface with error",
			mockFunc: func(m *mocks.MockManager) {
				m.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error"))
			},
			testCall: func(s *MessagingServer) (*api.AddResponse, error) {
				return s.AddBridge(context.Background(), &api.AddCommand{
					Name:   "test",
					Iface1: 1,
					Iface2: 2,
				})
			},
			wantErr: true,
		},
		{
			name: "AddBridge() without monitorIface without error",
			mockFunc: func(m *mocks.MockManager) {
				m.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			testCall: func(s *MessagingServer) (*api.AddResponse, error) {
				return s.AddBridge(context.Background(), &api.AddCommand{
					Name:   "test",
					Iface1: 1,
					Iface2: 2,
				})
			},
			wantErr: false,
		},
		{
			name: "AddBridge() with monitorIface with error",
			mockFunc: func(m *mocks.MockManager) {
				m.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error"))
			},
			testCall: func(s *MessagingServer) (*api.AddResponse, error) {
				return s.AddBridge(context.Background(), &api.AddCommand{
					Name:    "test",
					Iface1:  1,
					Iface2:  2,
					Monitor: new(int32),
				})
			},
			wantErr: true,
		},
		{
			name: "AddBridge() with monitorIface without error",
			mockFunc: func(m *mocks.MockManager) {
				m.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			testCall: func(s *MessagingServer) (*api.AddResponse, error) {
				return s.AddBridge(context.Background(), &api.AddCommand{
					Name:    "test",
					Iface1:  1,
					Iface2:  2,
					Monitor: new(int32),
				})
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			bridgeManager := mocks.NewMockManager(ctrl)

			server := NewMessagingServer(bridgeManager)
			if server == nil {
				t.Error("NewMessagingServer() returned nil")
			}

			tt.mockFunc(bridgeManager)

			out, _ := tt.testCall(server)
			if tt.wantErr && out.Success != false {
				t.Error("AddBridge() returned unexpected success")
			}
		})
	}
}

func TestMessagingServer_RemoveBridge(t *testing.T) {
	tests := []struct {
		name     string
		mockFunc func(*mocks.MockManager)
		testCall func(*MessagingServer) (*api.RemoveResponse, error)
		wantErr  bool
	}{
		{
			name: "RemoveBridge() with error",
			mockFunc: func(m *mocks.MockManager) {
				m.EXPECT().Remove(gomock.Any()).Return(fmt.Errorf("error"))
			},
			testCall: func(s *MessagingServer) (*api.RemoveResponse, error) {
				return s.RemoveBridge(context.Background(), &api.RemoveCommand{
					Name: "test",
				})
			},
			wantErr: true,
		},
		{
			name: "RemoveBridge() without error",
			mockFunc: func(m *mocks.MockManager) {
				m.EXPECT().Remove(gomock.Any()).Return(nil)
			},
			testCall: func(s *MessagingServer) (*api.RemoveResponse, error) {
				return s.RemoveBridge(context.Background(), &api.RemoveCommand{
					Name: "test",
				})
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			bridgeManager := mocks.NewMockManager(ctrl)

			server := NewMessagingServer(bridgeManager)
			if server == nil {
				t.Error("NewMessagingServer() returned nil")
			}

			tt.mockFunc(bridgeManager)

			out, _ := tt.testCall(server)
			if tt.wantErr && out.Success != false {
				t.Error("RemoveBridge() returned unexpected success")
			}
		})
	}
}

func TestMessagingServer_ListBridges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bridgeManager := mocks.NewMockManager(ctrl)
	server := NewMessagingServer(bridgeManager)
	if server == nil {
		t.Error("NewMessagingServer() returned nil")
	}

	bridgeManager.EXPECT().List().Return(nil)

	_, err := server.ListBridges(context.Background(), &api.ListCommand{})
	if err != nil {
		t.Errorf("ListBridges() error = %v", err)
	}
}
