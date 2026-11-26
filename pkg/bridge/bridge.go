// Package bridge implements a single eBPF bridge between two interfaces.
package bridge

//go:generate go tool mockgen -source=bridge.go -destination=mocks/bridge_mock.go -package=mocks

import (
	"errors"
	"fmt"

	"github.com/jklaiber/ebpf-bridge/pkg/bpf"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/linker"
)

type Bridge interface {
	Add() error
	Remove() error
	Name() string
	Interface1() hostlink.Link
	Interface2() hostlink.Link
	MonitorInterface() hostlink.Link
}

type BridgeFactory interface {
	NewBridge(name string, iface1 hostlink.Link, iface2 hostlink.Link, monitorIface hostlink.Link) Bridge
}

type EbpfBridgeFactory struct {
	loader    bpf.Loader
	newLinker linker.Factory
}

func NewEbpfBridgeFactory(loader bpf.Loader, newLinker linker.Factory) *EbpfBridgeFactory {
	return &EbpfBridgeFactory{
		loader:    loader,
		newLinker: newLinker,
	}
}

func (f *EbpfBridgeFactory) NewBridge(name string, iface1 hostlink.Link, iface2 hostlink.Link, monitorIface hostlink.Link) Bridge {
	return &EbpfBridge{
		loader:       f.loader,
		newLinker:    f.newLinker,
		name:         name,
		iface1:       iface1,
		iface2:       iface2,
		monitorIface: monitorIface,
	}
}

// EbpfBridge forwards every frame received on one interface to the other one
// and, if set, to the monitor interface. Frames received on the monitor
// interface are not forwarded.
type EbpfBridge struct {
	loader       bpf.Loader
	newLinker    linker.Factory
	name         string
	iface1       hostlink.Link
	iface2       hostlink.Link
	monitorIface hostlink.Link

	objects bpf.Objects
	linkers []linker.Linker
}

func (e *EbpfBridge) Name() string {
	return e.name
}

func (e *EbpfBridge) Interface1() hostlink.Link {
	return e.iface1
}

func (e *EbpfBridge) Interface2() hostlink.Link {
	return e.iface2
}

func (e *EbpfBridge) MonitorInterface() hostlink.Link {
	return e.monitorIface
}

// Add loads the program, configures the ports and attaches the program to
// both bridged interfaces. On failure everything done so far is rolled back.
func (e *EbpfBridge) Add() (err error) {
	if e.objects != nil {
		return fmt.Errorf("bridge %s is already active", e.name)
	}
	objects, err := e.loader.Load()
	if err != nil {
		return err
	}
	e.objects = objects
	defer func() {
		if err != nil {
			err = errors.Join(err, e.teardown())
		}
	}()

	if err := objects.SetPort(bpf.SlotIface1, e.iface1.Index()); err != nil {
		return err
	}
	if err := objects.SetPort(bpf.SlotIface2, e.iface2.Index()); err != nil {
		return err
	}
	if e.monitorIface != nil {
		if err := objects.SetPort(bpf.SlotMonitor, e.monitorIface.Index()); err != nil {
			return err
		}
	}

	for _, iface := range []hostlink.Link{e.iface1, e.iface2} {
		l := e.newLinker(iface, objects.Program())
		if err := l.Attach(); err != nil {
			return err
		}
		e.linkers = append(e.linkers, l)
	}
	return nil
}

// Remove detaches the program from all interfaces and releases it.
func (e *EbpfBridge) Remove() error {
	if e.objects == nil {
		return fmt.Errorf("bridge %s is not active", e.name)
	}
	if err := e.teardown(); err != nil {
		return fmt.Errorf("failed to remove bridge %s: %w", e.name, err)
	}
	return nil
}

func (e *EbpfBridge) teardown() error {
	var errs []error
	for _, l := range e.linkers {
		errs = append(errs, l.Detach())
	}
	e.linkers = nil
	if e.objects != nil {
		errs = append(errs, e.objects.Close())
		e.objects = nil
	}
	return errors.Join(errs...)
}
