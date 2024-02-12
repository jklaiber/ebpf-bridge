package command

import (
	"fmt"
	"testing"

	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	messagingMock "github.com/jklaiber/ebpf-bridge/pkg/messaging/mocks"
	"go.uber.org/mock/gomock"
)

func TestRemoveCommand_New(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessagingClient := messagingMock.NewMockClient(ctrl)

	rc := NewRemoveCommand(mockMessagingClient, "test")
	if rc == nil {
		t.Error("NewRemoveCommand returned nil")
	}
}

func TestRemoveCommand_Execute(t *testing.T) {
	tests := []struct {
		name     string
		mockCall func(ctrl *gomock.Controller, mockMessagingClient *messagingMock.MockClient)
		wantErr  bool
	}{
		{
			name: "Test Execute with message error",
			mockCall: func(ctrl *gomock.Controller, mockMessagingClient *messagingMock.MockClient) {
				mockMessagingClient.EXPECT().RemoveBridge(gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "Test Execute with no error",
			mockCall: func(ctrl *gomock.Controller, mockMessagingClient *messagingMock.MockClient) {
				mockMessagingClient.EXPECT().RemoveBridge(gomock.Any()).Return(&messaging.RemoveResponse{}, nil).AnyTimes()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMessagingClient := messagingMock.NewMockClient(ctrl)

			tt.mockCall(ctrl, mockMessagingClient)

			rc := NewRemoveCommand(mockMessagingClient, "test")
			_, err := rc.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
