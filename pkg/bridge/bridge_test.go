package bridge

import (
	"errors"
	"testing"

	"github.com/cilium/ebpf"

	"github.com/jklaiber/ebpf-bridge/pkg/bpf"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/linker"
)

type fakeLink struct{ index int }

func (l fakeLink) Index() int   { return l.index }
func (l fakeLink) Name() string { return "eth" }

// fakeObjects records the configured ports.
type fakeObjects struct {
	ports    map[uint32]int
	setErrAt int // fail the n-th SetPort call (1-based), 0 = never
	calls    int
	closed   bool
}

func (o *fakeObjects) Program() *ebpf.Program { return nil }
func (o *fakeObjects) SetPort(slot uint32, ifindex int) error {
	o.calls++
	if o.calls == o.setErrAt {
		return errors.New("map update failed")
	}
	o.ports[slot] = ifindex
	return nil
}
func (o *fakeObjects) Close() error { o.closed = true; return nil }

type fakeLoader struct {
	objs *fakeObjects
	err  error
}

func (l *fakeLoader) Load() (bpf.Objects, error) {
	if l.err != nil {
		return nil, l.err
	}
	return l.objs, nil
}

// fakeLinker records attach state per interface index.
type fakeLinker struct {
	iface    hostlink.Link
	attached map[int]bool
	fail     map[int]bool
	detachE  error
}

func (l *fakeLinker) Attach() error {
	if l.fail[l.iface.Index()] {
		return errors.New("attach failed")
	}
	l.attached[l.iface.Index()] = true
	return nil
}

func (l *fakeLinker) Detach() error {
	delete(l.attached, l.iface.Index())
	return l.detachE
}

type env struct {
	objs     *fakeObjects
	loader   *fakeLoader
	attached map[int]bool
	fail     map[int]bool
	detachE  error
}

func newEnv() *env {
	objs := &fakeObjects{ports: map[uint32]int{}}
	return &env{objs: objs, loader: &fakeLoader{objs: objs}, attached: map[int]bool{}, fail: map[int]bool{}}
}

func (e *env) bridge(monitor hostlink.Link) Bridge {
	factory := NewEbpfBridgeFactory(e.loader, linker.Factory(func(iface hostlink.Link, _ *ebpf.Program) linker.Linker {
		return &fakeLinker{iface: iface, attached: e.attached, fail: e.fail, detachE: e.detachE}
	}))
	return factory.NewBridge("br0", fakeLink{1}, fakeLink{2}, monitor)
}

func TestAddAndRemove(t *testing.T) {
	e := newEnv()
	b := e.bridge(fakeLink{3})
	if b.Name() != "br0" || b.Interface1().Index() != 1 || b.Interface2().Index() != 2 || b.MonitorInterface().Index() != 3 {
		t.Fatal("unexpected bridge accessors")
	}
	if err := b.Add(); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	want := map[uint32]int{bpf.SlotIface1: 1, bpf.SlotIface2: 2, bpf.SlotMonitor: 3}
	for slot, ifindex := range want {
		if e.objs.ports[slot] != ifindex {
			t.Errorf("port %d = %d, want %d", slot, e.objs.ports[slot], ifindex)
		}
	}
	if !e.attached[1] || !e.attached[2] || e.attached[3] {
		t.Errorf("attached = %v, want only 1 and 2", e.attached)
	}
	if err := b.Add(); err == nil {
		t.Error("second Add() succeeded, want error")
	}

	if err := b.Remove(); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if len(e.attached) != 0 || !e.objs.closed {
		t.Errorf("after Remove: attached=%v closed=%v", e.attached, e.objs.closed)
	}
	if err := b.Remove(); err == nil {
		t.Error("second Remove() succeeded, want error")
	}
}

func TestAddWithoutMonitor(t *testing.T) {
	e := newEnv()
	if err := e.bridge(nil).Add(); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if _, ok := e.objs.ports[bpf.SlotMonitor]; ok {
		t.Error("monitor port set without monitor interface")
	}
}

func TestAddRollsBack(t *testing.T) {
	tests := []struct {
		name  string
		setup func(e *env)
	}{
		{"load fails", func(e *env) { e.loader.err = errors.New("verifier") }},
		{"iface1 port fails", func(e *env) { e.objs.setErrAt = 1 }},
		{"iface2 port fails", func(e *env) { e.objs.setErrAt = 2 }},
		{"monitor port fails", func(e *env) { e.objs.setErrAt = 3 }},
		{"iface1 attach fails", func(e *env) { e.fail[1] = true }},
		{"iface2 attach fails", func(e *env) { e.fail[2] = true }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			tt.setup(e)
			b := e.bridge(fakeLink{3})
			if err := b.Add(); err == nil {
				t.Fatal("Add() succeeded, want error")
			}
			if len(e.attached) != 0 {
				t.Errorf("interfaces left attached: %v", e.attached)
			}
			if e.loader.err == nil && !e.objs.closed {
				t.Error("objects not closed after failed Add")
			}
			// A failed Add must leave the bridge reusable.
			e.loader.err, e.objs.setErrAt, e.objs.calls, e.fail = nil, 0, 0, map[int]bool{}
			if err := b.Add(); err != nil {
				t.Errorf("Add() after rollback error = %v", err)
			}
		})
	}
}

func TestRemoveReportsDetachErrors(t *testing.T) {
	e := newEnv()
	e.detachE = errors.New("detach failed")
	b := e.bridge(nil)
	if err := b.Add(); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := b.Remove(); err == nil {
		t.Error("Remove() succeeded, want error")
	}
	if !e.objs.closed {
		t.Error("objects not closed although detach failed")
	}
}
