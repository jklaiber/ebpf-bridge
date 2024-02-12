package command

import (
	"fmt"
	"testing"

	hostlinkMock "github.com/jklaiber/ebpf-bridge/pkg/hostlink/mocks"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	messagingMock "github.com/jklaiber/ebpf-bridge/pkg/messaging/mocks"
	"go.uber.org/mock/gomock"
)

func TestListCommand_New(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLinkFactory := hostlinkMock.NewMockLinkFactory(ctrl)
	mockMessagingClient := messagingMock.NewMockClient(ctrl)

	lc := NewListCommand(mockLinkFactory, mockMessagingClient)
	if lc == nil {
		t.Error("NewListCommand returned nil")
	}
}

func TestListCommand_Execute(t *testing.T) {
	tests := []struct {
		name     string
		mockCall func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient)
		wantErr  bool
	}{
		{
			name: "Test Execute with message error",
			mockCall: func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) {
				mockMessagingClient.EXPECT().ListBridges(&messaging.ListCommand{}).Return(nil, fmt.Errorf("error")).AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "Test Execute with no error",
			mockCall: func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockMessagingClient *messagingMock.MockClient) {
				mockMessagingClient.EXPECT().ListBridges(&messaging.ListCommand{}).Return(&messaging.ListResponse{}, nil).AnyTimes()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLinkFactory := hostlinkMock.NewMockLinkFactory(ctrl)
			mockMessagingClient := messagingMock.NewMockClient(ctrl)

			tt.mockCall(ctrl, mockLinkFactory, mockMessagingClient)

			lc := NewListCommand(mockLinkFactory, mockMessagingClient)
			_, err := lc.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
