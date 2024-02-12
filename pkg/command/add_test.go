package command

import (
	"fmt"
	"testing"

	hostlinkMock "github.com/jklaiber/ebpf-bridge/pkg/hostlink/mocks"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	messagingMock "github.com/jklaiber/ebpf-bridge/pkg/messaging/mocks"
	"go.uber.org/mock/gomock"
)

func TestAddCommand_New(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHostLink := hostlinkMock.NewMockLinkFactory(ctrl)
	mockMessagingClient := messagingMock.NewMockClient(ctrl)

	ac := NewAddCommand(mockHostLink, mockMessagingClient, "test", "iface1", "iface2", "monitorIface")
	if ac == nil {
		t.Error("NewAddCommand returned nil")
	}
}

func TestAddCommand_Execute(t *testing.T) {
	tests := []struct {
		name     string
		mockCall func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient)
		addCall  func(*hostlinkMock.MockLinkFactory, *messagingMock.MockClient) *AddCommand
		wantErr  bool
	}{
		{
			name: "Test Execute with iface1 error",
			mockCall: func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) {
				mockLinkFactory.EXPECT().NewLinkWithName("iface1").Return(nil, fmt.Errorf("error")).AnyTimes()
			},
			addCall: func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) *AddCommand {
				return NewAddCommand(mockLinkFactory, mockMessagingClient, "test", "iface1", "iface2", "monitorIface")
			},
			wantErr: true,
		},
		{
			name: "Test Execute with iface2 error",
			mockCall: func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) {
				mockLinkFactory.EXPECT().NewLinkWithName("iface1").Return(nil, nil).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("iface2").Return(nil, fmt.Errorf("error")).AnyTimes()
			},
			addCall: func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) *AddCommand {
				return NewAddCommand(mockLinkFactory, mockMessagingClient, "test", "iface1", "iface2", "monitorIface")
			},
			wantErr: true,
		},
		{
			name: "Test Execute with monitorIface error",
			mockCall: func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) {
				mockLinkFactory.EXPECT().NewLinkWithName("iface1").Return(nil, nil).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("iface2").Return(nil, nil).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("monitorIface").Return(nil, fmt.Errorf("error")).AnyTimes()
			},
			addCall: func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) *AddCommand {
				return NewAddCommand(mockLinkFactory, mockMessagingClient, "test", "iface1", "iface2", "monitorIface")
			},
			wantErr: true,
		},
		{
			name: "Test Execute with monitorIface with messaging error",
			mockCall: func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) {
				iface1Mock := hostlinkMock.NewMockLink(ctrl)
				iface1Mock.EXPECT().Index().Return(1).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("iface1").Return(iface1Mock, nil).AnyTimes()
				iface2Mock := hostlinkMock.NewMockLink(ctrl)
				iface2Mock.EXPECT().Index().Return(2).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("iface2").Return(iface2Mock, nil).AnyTimes()
				monitorIfaceMock := hostlinkMock.NewMockLink(ctrl)
				monitorIfaceMock.EXPECT().Index().Return(3).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("monitorIface").Return(monitorIfaceMock, nil).AnyTimes()
				mockMessagingClient.EXPECT().AddBridge(gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes()
			},
			addCall: func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) *AddCommand {
				return NewAddCommand(mockLinkFactory, mockMessagingClient, "test", "iface1", "iface2", "monitorIface")
			},
			wantErr: true,
		},
		{
			name: "Test Execute with monitorIface without error",
			mockCall: func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) {
				iface1Mock := hostlinkMock.NewMockLink(ctrl)
				iface1Mock.EXPECT().Index().Return(1).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("iface1").Return(iface1Mock, nil).AnyTimes()
				iface2Mock := hostlinkMock.NewMockLink(ctrl)
				iface2Mock.EXPECT().Index().Return(2).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("iface2").Return(iface2Mock, nil).AnyTimes()
				monitorIfaceMock := hostlinkMock.NewMockLink(ctrl)
				monitorIfaceMock.EXPECT().Index().Return(3).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("monitorIface").Return(monitorIfaceMock, nil).AnyTimes()
				mockMessagingClient.EXPECT().AddBridge(gomock.Any()).Return(&messaging.AddResponse{}, nil).AnyTimes()
			},
			addCall: func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) *AddCommand {
				return NewAddCommand(mockLinkFactory, mockMessagingClient, "test", "iface1", "iface2", "monitorIface")
			},
			wantErr: false,
		},
		{
			name: "Test Execute without monitorIface without error",
			mockCall: func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) {
				iface1Mock := hostlinkMock.NewMockLink(ctrl)
				iface1Mock.EXPECT().Index().Return(1).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("iface1").Return(iface1Mock, nil).AnyTimes()
				iface2Mock := hostlinkMock.NewMockLink(ctrl)
				iface2Mock.EXPECT().Index().Return(2).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("iface2").Return(iface2Mock, nil).AnyTimes()
				mockMessagingClient.EXPECT().AddBridge(gomock.Any()).Return(&messaging.AddResponse{}, nil).AnyTimes()
			},
			addCall: func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) *AddCommand {
				return NewAddCommand(mockLinkFactory, mockMessagingClient, "test", "iface1", "iface2", "")
			},
			wantErr: false,
		},
		{
			name: "Test Execute without monitorIface with messaging error",
			mockCall: func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) {
				iface1Mock := hostlinkMock.NewMockLink(ctrl)
				iface1Mock.EXPECT().Index().Return(1).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("iface1").Return(iface1Mock, nil).AnyTimes()
				iface2Mock := hostlinkMock.NewMockLink(ctrl)
				iface2Mock.EXPECT().Index().Return(2).AnyTimes()
				mockLinkFactory.EXPECT().NewLinkWithName("iface2").Return(iface2Mock, nil).AnyTimes()
				mockMessagingClient.EXPECT().AddBridge(gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes()
			},
			addCall: func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) *AddCommand {
				return NewAddCommand(mockLinkFactory, mockMessagingClient, "test", "iface1", "iface2", "")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLinkFactory := hostlinkMock.NewMockLinkFactory(ctrl)
			mockMessagingClient := messagingMock.NewMockClient(ctrl)

			ac := tt.addCall(mockLinkFactory, mockMessagingClient)
			tt.mockCall(ctrl, mockLinkFactory, mockMessagingClient)
			_, err := ac.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("AddCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

}
