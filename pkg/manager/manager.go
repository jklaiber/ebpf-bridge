//go:generate mockgen -source=manager.go -destination=mocks/manager_mock.go -package=mocks Manager
package manager

import (
	"fmt"

	"github.com/jklaiber/ebpf-bridge/pkg/bridge"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
)

type Manager interface {
	Add(name string, iface1 int, iface2 int, monitorIface *int) error
	Remove(name string) error
	List() []BridgeDescription
}

type BridgeManager struct {
	bridges       map[string]bridge.Bridge
	linkFactory   hostlink.LinkFactory
	bridgeFactory bridge.BridgeFactory
}

func NewBridgeManager(linkFactory hostlink.LinkFactory, bridgeFactory bridge.BridgeFactory) *BridgeManager {
	return &BridgeManager{
		bridges:       make(map[string]bridge.Bridge),
		linkFactory:   linkFactory,
		bridgeFactory: bridgeFactory,
	}
}

func (b *BridgeManager) Add(name string, iface1 int, iface2 int, monitorIface *int) error {
	niface1, err := b.linkFactory.NewLinkWithIndex(iface1)
	if err != nil {
		return fmt.Errorf("failed to get iface1: %w", err)
	}
	niface2, err := b.linkFactory.NewLinkWithIndex(iface2)
	if err != nil {
		return fmt.Errorf("failed to get iface2: %w", err)
	}
	if monitorIface != nil {
		nmonitorIface, err := b.linkFactory.NewLinkWithIndex(*monitorIface)
		if err != nil {
			return fmt.Errorf("failed to get monitorIface: %w", err)
		}
		ebpfBridge := b.bridgeFactory.NewBridge(name, niface1, niface2, nmonitorIface)
		err = ebpfBridge.Add()
		if err != nil {
			return err
		}
		b.bridges[name] = ebpfBridge
	} else {
		ebpfBridge := b.bridgeFactory.NewBridge(name, niface1, niface2, nil)
		err = ebpfBridge.Add()
		if err != nil {
			return err
		}
		b.bridges[name] = ebpfBridge
	}
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

type BridgeDescription struct {
	Name    string
	Iface1  int32
	Iface2  int32
	Monitor *int32
}

func (b *BridgeManager) List() []BridgeDescription {
	var bridges []BridgeDescription
	for name, bridge := range b.bridges {
		if bridge.MonitorInterface() == nil {
			bridges = append(bridges, BridgeDescription{
				Name:   name,
				Iface1: int32(bridge.Interface1().Index()),
				Iface2: int32(bridge.Interface2().Index()),
			})
		} else {
			monitorIndex := int32(bridge.MonitorInterface().Index())
			bridges = append(bridges, BridgeDescription{
				Name:    name,
				Iface1:  int32(bridge.Interface1().Index()),
				Iface2:  int32(bridge.Interface2().Index()),
				Monitor: &monitorIndex,
			})
		}
	}
	return bridges
}
