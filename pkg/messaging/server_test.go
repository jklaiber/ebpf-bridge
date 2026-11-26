package messaging

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
	"github.com/jklaiber/ebpf-bridge/pkg/manager"
	"github.com/jklaiber/ebpf-bridge/pkg/manager/mocks"
)

func ptr[T any](v T) *T { return &v }

func TestServerAddBridge(t *testing.T) {
	tests := []struct {
		name        string
		in          *api.AddCommand
		wantMonitor *int
		managerErr  error
		wantSuccess bool
	}{
		{"without monitor", &api.AddCommand{Name: "br", Iface1: 1, Iface2: 2}, nil, nil, true},
		{"with monitor", &api.AddCommand{Name: "br", Iface1: 1, Iface2: 2, Monitor: ptr[int32](3)}, ptr(3), nil, true},
		{"manager error", &api.AddCommand{Name: "br", Iface1: 1, Iface2: 2}, nil, errors.New("boom"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks.NewMockManager(gomock.NewController(t))
			m.EXPECT().Add("br", 1, 2, gomock.Eq(tt.wantMonitor)).Return(tt.managerErr)

			resp, err := NewMessagingServer(m, "").AddBridge(context.Background(), tt.in)
			if err != nil {
				t.Fatalf("AddBridge() error = %v", err)
			}
			if resp.GetSuccess() != tt.wantSuccess {
				t.Errorf("Success = %v, want %v (%s)", resp.GetSuccess(), tt.wantSuccess, resp.GetMessage())
			}
			if resp.GetMessage() == "" {
				t.Error("empty message")
			}
		})
	}
}

func TestServerRemoveBridge(t *testing.T) {
	for _, managerErr := range []error{nil, errors.New("boom")} {
		m := mocks.NewMockManager(gomock.NewController(t))
		m.EXPECT().Remove("br").Return(managerErr)

		resp, err := NewMessagingServer(m, "").RemoveBridge(context.Background(), &api.RemoveCommand{Name: "br"})
		if err != nil {
			t.Fatalf("RemoveBridge() error = %v", err)
		}
		if resp.GetSuccess() != (managerErr == nil) {
			t.Errorf("Success = %v with manager error %v", resp.GetSuccess(), managerErr)
		}
	}
}

func TestServerListBridges(t *testing.T) {
	m := mocks.NewMockManager(gomock.NewController(t))
	m.EXPECT().List().Return([]manager.BridgeDescription{
		{Name: "a", Iface1: 1, Iface2: 2},
		{Name: "b", Iface1: 3, Iface2: 4, Monitor: ptr[int32](5)},
	})

	resp, err := NewMessagingServer(m, "").ListBridges(context.Background(), &api.ListCommand{})
	if err != nil {
		t.Fatalf("ListBridges() error = %v", err)
	}
	got := resp.GetBridges()
	if len(got) != 2 || got[0].GetName() != "a" || got[0].Monitor != nil || got[1].GetMonitor() != 5 {
		t.Errorf("unexpected bridges %v", got)
	}
}

func TestServerStartFailsOnBadSocket(t *testing.T) {
	m := mocks.NewMockManager(gomock.NewController(t))
	// The parent of the socket is a regular file, so it cannot be created.
	file := t.TempDir() + "/file"
	if err := writeFile(file); err != nil {
		t.Fatal(err)
	}
	if err := NewMessagingServer(m, file+"/sock").Start(); err == nil {
		t.Error("Start() succeeded, want error")
	}
}
