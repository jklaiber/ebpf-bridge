package messaging

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
	apiMocks "github.com/jklaiber/ebpf-bridge/pkg/api/mocks"
	"github.com/jklaiber/ebpf-bridge/pkg/manager"
	"github.com/jklaiber/ebpf-bridge/pkg/manager/mocks"
)

func writeFile(path string) error {
	return os.WriteFile(path, nil, 0o600)
}

// shortSocket returns a socket path that fits into sun_path on every OS
// (macOS limits it to 104 bytes, t.TempDir() can be longer).
func shortSocket(t *testing.T) string {
	dir, err := os.MkdirTemp("", "ebb")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return filepath.Join(dir, "s.sock")
}

func TestClientCallsController(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := apiMocks.NewMockEbpfBridgeControllerClient(ctrl)
	mc := &MessagingClient{client: c, timeout: time.Second}

	add := &api.AddCommand{Name: "br", Iface1: 1, Iface2: 2}
	c.EXPECT().AddBridge(gomock.Any(), add).Return(&api.AddResponse{Success: true}, nil)
	remove := &api.RemoveCommand{Name: "br"}
	c.EXPECT().RemoveBridge(gomock.Any(), remove).Return(&api.RemoveResponse{Success: true}, nil)
	list := &api.ListCommand{}
	c.EXPECT().ListBridges(gomock.Any(), list).Return(nil, errors.New("boom"))

	if _, err := mc.AddBridge(add); err != nil {
		t.Errorf("AddBridge() error = %v", err)
	}
	if _, err := mc.RemoveBridge(remove); err != nil {
		t.Errorf("RemoveBridge() error = %v", err)
	}
	if _, err := mc.ListBridges(list); err == nil || errors.Is(err, ErrServiceUnavailable) {
		t.Errorf("ListBridges() error = %v, want the controller error", err)
	}
	if err := mc.Close(); err != nil {
		t.Errorf("Close() without connection error = %v", err)
	}
}

// TestRoundTrip runs client and server over a real unix socket.
func TestRoundTrip(t *testing.T) {
	socket := shortSocket(t)
	m := mocks.NewMockManager(gomock.NewController(t))
	server := NewMessagingServer(m, socket)
	if err := server.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer server.Stop()

	info, err := os.Stat(socket)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != socketMode {
		t.Errorf("socket mode = %o, want %o", perm, socketMode)
	}

	client, err := NewMessagingClient(socket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	m.EXPECT().Add("br", 1, 2, gomock.Eq(ptr(3))).Return(nil)
	resp, err := client.AddBridge(&api.AddCommand{Name: "br", Iface1: 1, Iface2: 2, Monitor: ptr[int32](3)})
	if err != nil || !resp.GetSuccess() {
		t.Fatalf("AddBridge() = %v, %v", resp, err)
	}

	m.EXPECT().List().Return([]manager.BridgeDescription{{Name: "br", Iface1: 1, Iface2: 2}})
	list, err := client.ListBridges(&api.ListCommand{})
	if err != nil || len(list.GetBridges()) != 1 {
		t.Fatalf("ListBridges() = %v, %v", list, err)
	}

	m.EXPECT().Remove("br").Return(errors.New("busy"))
	rm, err := client.RemoveBridge(&api.RemoveCommand{Name: "br"})
	if err != nil || rm.GetSuccess() || !strings.Contains(rm.GetMessage(), "busy") {
		t.Fatalf("RemoveBridge() = %v, %v", rm, err)
	}
}

func TestStartReplacesStaleSocket(t *testing.T) {
	socket := shortSocket(t)
	if err := writeFile(socket); err != nil {
		t.Fatal(err)
	}
	server := NewMessagingServer(mocks.NewMockManager(gomock.NewController(t)), socket)
	if err := server.Start(); err != nil {
		t.Fatalf("Start() with stale socket error = %v", err)
	}
	server.Stop()
}

func TestClientServiceUnavailable(t *testing.T) {
	client, err := NewMessagingClient(shortSocket(t))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	_, err = client.ListBridges(&api.ListCommand{})
	if !errors.Is(err, ErrServiceUnavailable) {
		t.Errorf("error = %v, want ErrServiceUnavailable", err)
	}
}
