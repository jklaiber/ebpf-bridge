package bridge

import (
	"fmt"

	"github.com/jklaiber/ebpf-bridge/pkg/bpf"
	"github.com/jklaiber/ebpf-bridge/pkg/linker"
	"github.com/vishvananda/netlink"
)

type Bridge interface {
	Add(name string, iface1 netlink.Link, iface2 netlink.Link, monitorIface netlink.Link) error
	Remove(name string) error
}

type EbpfBridge struct {
	Name         string
	iface1       netlink.Link
	iface2       netlink.Link
	monitorIface netlink.Link
	iface1Linker linker.Linker
	iface2Linker linker.Linker
}

func NewEbpfBridge(name string, iface1 netlink.Link, iface2 netlink.Link, monitorIface netlink.Link) *EbpfBridge {
	return &EbpfBridge{
		Name:         name,
		iface1:       iface1,
		iface2:       iface2,
		monitorIface: monitorIface,
	}
}

func (e *EbpfBridge) Add() error {
	realBpf := &bpf.BpfLinux{}
	bpfObjects, err := realBpf.ReadBpfObjects()
	if err != nil {
		return fmt.Errorf("failed to read bpf objects: %w", err)
	}
	linkerIface1 := linker.NewXdpLinker(e.iface1, bpfObjects.XdpBridge)
	e.iface1Linker = linkerIface1
	linkerIface2 := linker.NewXdpLinker(e.iface2, bpfObjects.XdpBridge)
	e.iface2Linker = linkerIface2

	if err := bpfObjects.Devmap.Put(uint32(0), uint32(e.iface1.Attrs().Index)); err != nil {
		return fmt.Errorf("failed to put iface1 into devmap: %w", err)
	}
	if err := bpfObjects.Devmap.Put(uint32(1), uint32(e.iface2.Attrs().Index)); err != nil {
		return fmt.Errorf("failed to put iface2 into devmap: %w", err)
	}
	if err := bpfObjects.Devmap.Put(uint32(2), uint32(e.monitorIface.Attrs().Index)); err != nil {
		return fmt.Errorf("failed to put monitorIface into devmap: %w", err)
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
	return nil
}
