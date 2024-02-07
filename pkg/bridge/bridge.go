package bridge

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jklaiber/ebpf-bridge/pkg/bpf"
	"github.com/jklaiber/ebpf-bridge/pkg/linker"
	"github.com/vishvananda/netlink"
)

const PinPath = "/sys/fs/bpf/devmap_"

type Bridge interface {
	Add(name string, iface1 netlink.Link, iface2 netlink.Link, monitorIface netlink.Link) error
	Remove(name string) error
}

type EbpfBridge struct {
	bpf          bpf.Bpf
	Name         string
	Iface1       netlink.Link
	Iface2       netlink.Link
	MonitorIface netlink.Link
	Iface1Linker linker.Linker
	Iface2Linker linker.Linker
	mapUuid      string
}

func NewEbpfBridge(name string, iface1 netlink.Link, iface2 netlink.Link, monitorIface netlink.Link) *EbpfBridge {
	return &EbpfBridge{
		bpf:          &bpf.BpfLinux{},
		Name:         name,
		Iface1:       iface1,
		Iface2:       iface2,
		MonitorIface: monitorIface,
		mapUuid:      uuid.New().String(),
	}
}

func (e *EbpfBridge) Add() error {
	bpfObjects, err := e.bpf.ReadBpfObjects()
	if err != nil {
		return fmt.Errorf("failed to read bpf objects: %w", err)
	}
	linkerIface1 := linker.NewXdpLinker(e.Iface1, bpfObjects.XdpBridge)
	e.Iface1Linker = linkerIface1
	linkerIface2 := linker.NewXdpLinker(e.Iface2, bpfObjects.XdpBridge)
	e.Iface2Linker = linkerIface2

	err = bpfObjects.Devmap.Pin(PinPath + e.mapUuid)
	if err != nil {
		return fmt.Errorf("failed to pin devmap: %w", err)
	}

	if err := bpfObjects.Devmap.Put(uint32(0), uint32(e.Iface1.Attrs().Index)); err != nil {
		return fmt.Errorf("failed to put iface1 into devmap: %w", err)
	}
	if err := bpfObjects.Devmap.Put(uint32(1), uint32(e.Iface2.Attrs().Index)); err != nil {
		return fmt.Errorf("failed to put iface2 into devmap: %w", err)
	}
	if e.MonitorIface != nil {
		if err := bpfObjects.Devmap.Put(uint32(2), uint32(e.MonitorIface.Attrs().Index)); err != nil {
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
	err := e.Iface1Linker.Detach()
	if err != nil {
		return fmt.Errorf("failed to detach iface1: %w", err)
	}
	err = e.Iface2Linker.Detach()
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
