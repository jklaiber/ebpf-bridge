package bridge

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

type Manager interface {
	Add(name string, iface1 int, iface2 int, monitorIface *int) error
	Remove(name string) error
	List()
}

type BridgeManager struct {
	bridges map[string]*EbpfBridge
}

func NewBridgeManager() *BridgeManager {
	return &BridgeManager{
		bridges: make(map[string]*EbpfBridge),
	}
}

func (b *BridgeManager) Add(name string, iface1 int, iface2 int, monitorIface *int) error {
	niface1, err := netlink.LinkByIndex(iface1)
	if err != nil {
		return fmt.Errorf("failed to get iface1: %w", err)
	}
	niface2, err := netlink.LinkByIndex(iface2)
	if err != nil {
		return fmt.Errorf("failed to get iface2: %w", err)
	}
	if monitorIface != nil {
		nmonitorIface, err := netlink.LinkByIndex(*monitorIface)
		if err != nil {
			return fmt.Errorf("failed to get monitorIface: %w", err)
		}
		ebpfBridge := NewEbpfBridge(name, niface1, niface2, nmonitorIface)
		err = ebpfBridge.Add()
		if err != nil {
			return err
		}
		b.bridges[name] = ebpfBridge
	} else {
		ebpfBridge := NewEbpfBridge(name, niface1, niface2, nil)
		err = ebpfBridge.Add()
		if err != nil {
			return err
		}
		b.bridges[name] = ebpfBridge
	}
	// nmonitorIface, err := netlink.LinkByIndex(monitorIface)
	// if err != nil {
	// 	return fmt.Errorf("failed to get monitorIface: %w", err)
	// }
	// ebpfBridge := NewEbpfBridge(name, niface1, niface2, nmonitorIface)
	// ebpfBridge := NewEbpfBridge(name, niface1, niface2, nil)
	// err = ebpfBridge.Add()
	// if err != nil {
	// 	return err
	// }
	// b.bridges[name] = ebpfBridge
	return nil
}

func (b *BridgeManager) Remove(name string) error {
	ebpfBridge, ok := b.bridges[name]
	if !ok {
		return fmt.Errorf("bridge %s does not exist", name)
	}
	err := ebpfBridge.Remove()
	if err != nil {
		return fmt.Errorf("failed to remove bridge %s: %w", name, err)
	}
	delete(b.bridges, name)
	return nil
}

func (b *BridgeManager) List() {
	for name, _ := range b.bridges {
		println(name)
	}
}
