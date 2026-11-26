package command

import (
	"errors"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	hostlinkMock "github.com/jklaiber/ebpf-bridge/pkg/hostlink/mocks"
	messagingMock "github.com/jklaiber/ebpf-bridge/pkg/messaging/mocks"
)

type fakeLink struct {
	index int
	name  string
}

func (l fakeLink) Index() int   { return l.index }
func (l fakeLink) Name() string { return l.name }

var links = map[string]int{"eth1": 1, "eth2": 2, "eth3": 3}

// linkFactory resolves the interfaces in links.
func linkFactory(ctrl *gomock.Controller) *hostlinkMock.MockLinkFactory {
	f := hostlinkMock.NewMockLinkFactory(ctrl)
	f.EXPECT().NewLinkWithName(gomock.Any()).DoAndReturn(func(name string) (hostlink.Link, error) {
		if i, ok := links[name]; ok {
			return fakeLink{i, name}, nil
		}
		return nil, errors.New("Link not found")
	}).AnyTimes()
	f.EXPECT().NewLinkWithIndex(gomock.Any()).DoAndReturn(func(index int) (hostlink.Link, error) {
		for name, i := range links {
			if i == index {
				return fakeLink{i, name}, nil
			}
		}
		return nil, errors.New("Link not found")
	}).AnyTimes()
	return f
}

func ptr[T any](v T) *T { return &v }

func TestAddCommand(t *testing.T) {
	tests := []struct {
		name    string
		monitor string
		iface1  string
		want    *api.AddCommand
		resp    *api.AddResponse
		rpcErr  error
		wantErr string
	}{
		{
			name: "with monitor", iface1: "eth1", monitor: "eth3",
			want: &api.AddCommand{Name: "br", Iface1: 1, Iface2: 2, Monitor: ptr[int32](3)},
			resp: &api.AddResponse{Success: true, Message: "Bridge br added"},
		},
		{
			name: "without monitor", iface1: "eth1",
			want: &api.AddCommand{Name: "br", Iface1: 1, Iface2: 2},
			resp: &api.AddResponse{Success: true, Message: "Bridge br added"},
		},
		{name: "unknown iface1", iface1: "nope", wantErr: "interface nope"},
		{name: "unknown monitor", iface1: "eth1", monitor: "nope", wantErr: "interface nope"},
		{
			name: "service rejects", iface1: "eth1",
			want:    &api.AddCommand{Name: "br", Iface1: 1, Iface2: 2},
			resp:    &api.AddResponse{Success: false, Message: "bridge br already exists"},
			wantErr: "already exists",
		},
		{
			name: "rpc error", iface1: "eth1",
			want:    &api.AddCommand{Name: "br", Iface1: 1, Iface2: 2},
			rpcErr:  errors.New("unavailable"),
			wantErr: "unavailable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			client := messagingMock.NewMockClient(ctrl)
			if tt.want != nil {
				client.EXPECT().AddBridge(gomock.Eq(tt.want)).Return(tt.resp, tt.rpcErr)
			}
			msg, err := NewAddCommand(linkFactory(ctrl), client, "br", tt.iface1, "eth2", tt.monitor).Execute()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Execute() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || msg != tt.resp.GetMessage() {
				t.Errorf("Execute() = %q, %v", msg, err)
			}
		})
	}
}

func TestRemoveCommand(t *testing.T) {
	tests := []struct {
		name    string
		resp    *api.RemoveResponse
		rpcErr  error
		wantErr bool
	}{
		{"success", &api.RemoveResponse{Success: true, Message: "Bridge br removed"}, nil, false},
		{"service rejects", &api.RemoveResponse{Success: false, Message: "bridge br does not exist"}, nil, true},
		{"rpc error", nil, errors.New("unavailable"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := messagingMock.NewMockClient(gomock.NewController(t))
			client.EXPECT().RemoveBridge(gomock.Eq(&api.RemoveCommand{Name: "br"})).Return(tt.resp, tt.rpcErr)
			msg, err := NewRemoveCommand(client, "br").Execute()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && msg != tt.resp.GetMessage() {
				t.Errorf("Execute() = %q", msg)
			}
		})
	}
}

func TestListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := messagingMock.NewMockClient(ctrl)
	client.EXPECT().ListBridges(gomock.Any()).Return(&api.ListResponse{Bridges: []*api.BridgeDescription{
		{Name: "br", Iface1: 1, Iface2: 2, Monitor: ptr[int32](3)},
	}}, nil)
	out, err := NewListCommand(linkFactory(ctrl), client).Execute()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"br", "eth1", "eth2", "eth3"} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not contain %q:\n%s", want, out)
		}
	}

	client.EXPECT().ListBridges(gomock.Any()).Return(nil, errors.New("unavailable"))
	if _, err := NewListCommand(linkFactory(ctrl), client).Execute(); err == nil {
		t.Error("Execute() succeeded, want error")
	}
}
