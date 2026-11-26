package manager

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/jklaiber/ebpf-bridge/pkg/bridge"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	hostlinkMock "github.com/jklaiber/ebpf-bridge/pkg/hostlink/mocks"
)

// fakeLink is a hostlink.Link with a fixed index and name.
type fakeLink struct{ index int }

func (l fakeLink) Index() int    { return l.index }
func (l fakeLink) Name() string  { return fmt.Sprintf("eth%d", l.index) }
func link(index int) fakeLink    { return fakeLink{index: index} }
func ptr[T any](v T) *T          { return &v }
func idx(l hostlink.Link) string { return fmt.Sprint(l.Index()) }

// fakeBridge records Add and Remove calls.
type fakeBridge struct {
	name                    string
	iface1, iface2, monitor hostlink.Link
	addErr, removeErr       error
	added, removed          bool
}

func (b *fakeBridge) Add() error                      { b.added = true; return b.addErr }
func (b *fakeBridge) Remove() error                   { b.removed = true; return b.removeErr }
func (b *fakeBridge) Name() string                    { return b.name }
func (b *fakeBridge) Interface1() hostlink.Link       { return b.iface1 }
func (b *fakeBridge) Interface2() hostlink.Link       { return b.iface2 }
func (b *fakeBridge) MonitorInterface() hostlink.Link { return b.monitor }

type fakeBridgeFactory struct {
	mu        sync.Mutex
	bridges   []*fakeBridge
	addErr    error
	removeErr error
}

func (f *fakeBridgeFactory) NewBridge(name string, iface1, iface2, monitor hostlink.Link) bridge.Bridge {
	f.mu.Lock()
	defer f.mu.Unlock()
	b := &fakeBridge{name: name, iface1: iface1, iface2: iface2, monitor: monitor, addErr: f.addErr, removeErr: f.removeErr}
	f.bridges = append(f.bridges, b)
	return b
}

// newManager returns a manager whose link factory resolves every index.
func newManager(t *testing.T) (*BridgeManager, *fakeBridgeFactory) {
	ctrl := gomock.NewController(t)
	links := hostlinkMock.NewMockLinkFactory(ctrl)
	links.EXPECT().NewLinkWithIndex(gomock.Any()).DoAndReturn(func(i int) (hostlink.Link, error) {
		if i >= 100 {
			return nil, errors.New("no such device")
		}
		return link(i), nil
	}).AnyTimes()
	bridges := &fakeBridgeFactory{}
	return NewBridgeManager(links, bridges), bridges
}

func TestAdd(t *testing.T) {
	m, factory := newManager(t)
	if err := m.Add("br0", 1, 2, ptr(3)); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if len(factory.bridges) != 1 {
		t.Fatalf("created %d bridges, want 1", len(factory.bridges))
	}
	b := factory.bridges[0]
	if !b.added || b.name != "br0" || idx(b.iface1) != "1" || idx(b.iface2) != "2" || idx(b.monitor) != "3" {
		t.Errorf("unexpected bridge %+v", b)
	}
}

func TestAddWithoutMonitor(t *testing.T) {
	m, factory := newManager(t)
	if err := m.Add("br0", 1, 2, nil); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if factory.bridges[0].monitor != nil {
		t.Error("monitor interface set, want nil")
	}
}

func TestAddErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *BridgeManager)
		bridge  string
		iface1  int
		iface2  int
		monitor *int
		wantErr string
	}{
		{name: "empty name", bridge: " ", iface1: 1, iface2: 2, wantErr: "must not be empty"},
		{name: "same interfaces", bridge: "x", iface1: 1, iface2: 1, wantErr: "only be used once"},
		{name: "monitor equals iface1", bridge: "x", iface1: 1, iface2: 2, monitor: ptr(1), wantErr: "only be used once"},
		{name: "monitor equals iface2", bridge: "x", iface1: 1, iface2: 2, monitor: ptr(2), wantErr: "only be used once"},
		{name: "unknown iface1", bridge: "x", iface1: 100, iface2: 2, wantErr: "iface1"},
		{name: "unknown iface2", bridge: "x", iface1: 1, iface2: 100, wantErr: "iface2"},
		{name: "unknown monitor", bridge: "x", iface1: 1, iface2: 2, monitor: ptr(100), wantErr: "monitorIface"},
		{
			name:   "duplicate name",
			setup:  func(m *BridgeManager) { _ = m.Add("x", 10, 11, nil) },
			bridge: "x", iface1: 1, iface2: 2,
			wantErr: "already exists",
		},
		{
			name:   "iface used by other bridge",
			setup:  func(m *BridgeManager) { _ = m.Add("other", 10, 11, nil) },
			bridge: "x", iface1: 11, iface2: 2,
			wantErr: "already used by bridge other",
		},
		{
			name:   "monitor of other bridge",
			setup:  func(m *BridgeManager) { _ = m.Add("other", 10, 11, ptr(12)) },
			bridge: "x", iface1: 1, iface2: 2, monitor: ptr(12),
			wantErr: "already used by bridge other",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := newManager(t)
			if tt.setup != nil {
				tt.setup(m)
			}
			before := len(m.List())
			err := m.Add(tt.bridge, tt.iface1, tt.iface2, tt.monitor)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Add() error = %v, want it to contain %q", err, tt.wantErr)
			}
			if got := len(m.List()); got != before {
				t.Errorf("bridges after failed Add = %d, want %d", got, before)
			}
		})
	}
}

func TestAddBridgeFailureIsNotRegistered(t *testing.T) {
	m, factory := newManager(t)
	factory.addErr = errors.New("attach failed")
	if err := m.Add("br0", 1, 2, nil); err == nil {
		t.Fatal("Add() succeeded, want error")
	}
	if len(m.List()) != 0 {
		t.Error("failed bridge was registered")
	}
	factory.addErr = nil
	if err := m.Add("br0", 1, 2, nil); err != nil {
		t.Errorf("retrying Add() error = %v", err)
	}
}

func TestRemove(t *testing.T) {
	m, factory := newManager(t)
	_ = m.Add("br0", 1, 2, nil)
	if err := m.Remove("br0"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if !factory.bridges[0].removed {
		t.Error("bridge was not removed")
	}
	if len(m.List()) != 0 {
		t.Error("bridge still listed after Remove")
	}
	// The interfaces are free again.
	if err := m.Add("br1", 1, 2, nil); err != nil {
		t.Errorf("Add() after Remove error = %v", err)
	}
}

func TestRemoveUnknown(t *testing.T) {
	m, _ := newManager(t)
	if err := m.Remove("nope"); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Remove() error = %v, want 'does not exist'", err)
	}
}

func TestRemoveFailureKeepsBridge(t *testing.T) {
	m, factory := newManager(t)
	factory.removeErr = errors.New("detach failed")
	_ = m.Add("br0", 1, 2, nil)
	if err := m.Remove("br0"); err == nil {
		t.Fatal("Remove() succeeded, want error")
	}
	if len(m.List()) != 1 {
		t.Error("bridge dropped although Remove failed")
	}
}

func TestRemoveAll(t *testing.T) {
	m, factory := newManager(t)
	_ = m.Add("a", 1, 2, nil)
	_ = m.Add("b", 3, 4, nil)
	if err := m.RemoveAll(); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}
	for _, b := range factory.bridges {
		if !b.removed {
			t.Errorf("bridge %s not removed", b.name)
		}
	}
	if len(m.List()) != 0 {
		t.Error("bridges left after RemoveAll")
	}
}

func TestRemoveAllReportsErrors(t *testing.T) {
	m, factory := newManager(t)
	factory.removeErr = errors.New("detach failed")
	_ = m.Add("a", 1, 2, nil)
	if err := m.RemoveAll(); err == nil {
		t.Error("RemoveAll() succeeded, want error")
	}
}

func TestListIsSorted(t *testing.T) {
	m, _ := newManager(t)
	_ = m.Add("c", 5, 6, ptr(7))
	_ = m.Add("a", 1, 2, nil)
	_ = m.Add("b", 3, 4, nil)

	got := m.List()
	var names []string
	for _, b := range got {
		names = append(names, b.Name)
	}
	if strings.Join(names, ",") != "a,b,c" {
		t.Errorf("List() order = %v, want a,b,c", names)
	}
	if got[0].Monitor != nil {
		t.Error("bridge a has a monitor")
	}
	if got[2].Iface1 != 5 || got[2].Iface2 != 6 || got[2].Monitor == nil || *got[2].Monitor != 7 {
		t.Errorf("bridge c = %+v", got[2])
	}
}

func TestConcurrentUse(t *testing.T) {
	m, _ := newManager(t)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			name := fmt.Sprintf("br%d", i)
			_ = m.Add(name, 2*i+1, 2*i+2, nil)
			_ = m.List()
			if i%2 == 0 {
				_ = m.Remove(name)
			}
		}()
	}
	wg.Wait()
	if got := len(m.List()); got != 10 {
		t.Errorf("bridges = %d, want 10", got)
	}
}
