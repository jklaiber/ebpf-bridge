//go:generate mockgen -source=bridge.go -destination=mocks/bridge_mock.go -package=mocks Bridge
package bridge

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jklaiber/ebpf-bridge/pkg/bpf"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/linker"
)

const PinPath = "/sys/fs/bpf/devmap_"

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
	bpf bpf.Bpf
}

func NewEbpfBridgeFactory(bpf bpf.Bpf) *EbpfBridgeFactory {
	return &EbpfBridgeFactory{
		bpf: bpf,
	}
}

func (f *EbpfBridgeFactory) NewBridge(name string, iface1 hostlink.Link, iface2 hostlink.Link, monitorIface hostlink.Link) Bridge {
	return &EbpfBridge{
		bpf:          f.bpf,
		name:         name,
		iface1:       iface1,
		iface2:       iface2,
		monitorIface: monitorIface,
		mapUuid:      uuid.New().String(),
	}
}

type EbpfBridge struct {
	bpf          bpf.Bpf
	name         string
	iface1       hostlink.Link
	iface2       hostlink.Link
	monitorIface hostlink.Link
	iface1Linker linker.Linker
	iface2Linker linker.Linker
	mapUuid      string
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

func (e *EbpfBridge) Add() error {
	bpfObjects, err := e.bpf.ReadBpfObjects()
	if err != nil {
		return fmt.Errorf("failed to read bpf objects: %w", err)
	}
	linkerIface1 := linker.NewXdpLinker(e.iface1, bpfObjects.XdpBridge)
	e.iface1Linker = linkerIface1
	linkerIface2 := linker.NewXdpLinker(e.iface2, bpfObjects.XdpBridge)
	e.iface2Linker = linkerIface2

	err = bpfObjects.Devmap.Pin(PinPath + e.mapUuid)
	if err != nil {
		return fmt.Errorf("failed to pin devmap: %w", err)
	}

	if err := bpfObjects.Devmap.Put(uint32(0), uint32(e.iface1.Index())); err != nil {
		return fmt.Errorf("failed to put iface1 into devmap: %w", err)
	}
	if err := bpfObjects.Devmap.Put(uint32(1), uint32(e.iface2.Index())); err != nil {
		return fmt.Errorf("failed to put iface2 into devmap: %w", err)
	}
	if e.monitorIface != nil {
		if err := bpfObjects.Devmap.Put(uint32(2), uint32(e.monitorIface.Index())); err != nil {
			return fmt.Errorf("failed to put monitorIface into devmap: %w", err)
		}
	}
	err = linkerIface1.Attach()
	if err != nil {
		return fmt.Errorf("failed to attach iface1: %w", err)
	}
	err = linkerIface2.Attach()
	if err != nil {
		return fmt.Errorf("failed to attach iface2: %w", err)
	}

	return nil
}

func (e *EbpfBridge) Remove() error {
	err := e.iface1Linker.Detach()
	if err != nil {
		return fmt.Errorf("failed to detach iface1: %w", err)
	}
	err = e.iface2Linker.Detach()
	if err != nil {
		return fmt.Errorf("failed to detach iface2: %w", err)
	}

	m, err := e.bpf.LoadPinnedMap(PinPath + e.mapUuid)
	if err != nil {
		return fmt.Errorf("failed to load pinned map: %w", err)
	}
	err = m.Unpin()
	if err != nil {
		return fmt.Errorf("failed to unpin map: %w", err)
	}
	err = m.Close()
	if err != nil {
		return fmt.Errorf("failed to close map: %w", err)
	}

	return nil
}
