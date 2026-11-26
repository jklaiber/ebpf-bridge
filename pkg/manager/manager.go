// Package manager keeps track of all bridges managed by the service.
package manager

//go:generate go tool mockgen -source=manager.go -destination=mocks/manager_mock.go -package=mocks

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/jklaiber/ebpf-bridge/pkg/bridge"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
)

type Manager interface {
	Add(name string, iface1 int, iface2 int, monitorIface *int) error
	Remove(name string) error
	RemoveAll() error
	List() []BridgeDescription
}

type BridgeDescription struct {
	Name    string
	Iface1  int32
	Iface2  int32
	Monitor *int32
}

// BridgeManager is safe for concurrent use.
type BridgeManager struct {
	mu            sync.Mutex
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
	if strings.TrimSpace(name) == "" {
		return errors.New("bridge name must not be empty")
	}
	indexes := []int{iface1, iface2}
	if monitorIface != nil {
		indexes = append(indexes, *monitorIface)
	}
	for i, idx := range indexes {
		if slices.Contains(indexes[:i], idx) {
			return errors.New("an interface can only be used once per bridge")
		}
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.bridges[name]; ok {
		return fmt.Errorf("bridge %s already exists", name)
	}
	for _, idx := range indexes {
		if other, ok := b.bridgeUsing(idx); ok {
			return fmt.Errorf("interface with index %d is already used by bridge %s", idx, other)
		}
	}

	niface1, err := b.linkFactory.NewLinkWithIndex(iface1)
	if err != nil {
		return fmt.Errorf("failed to get iface1: %w", err)
	}
	niface2, err := b.linkFactory.NewLinkWithIndex(iface2)
	if err != nil {
		return fmt.Errorf("failed to get iface2: %w", err)
	}
	var nmonitor hostlink.Link
	if monitorIface != nil {
		nmonitor, err = b.linkFactory.NewLinkWithIndex(*monitorIface)
		if err != nil {
			return fmt.Errorf("failed to get monitorIface: %w", err)
		}
	}

	ebpfBridge := b.bridgeFactory.NewBridge(name, niface1, niface2, nmonitor)
	if err := ebpfBridge.Add(); err != nil {
		return err
	}
	b.bridges[name] = ebpfBridge
	return nil
}

// bridgeUsing returns the name of the bridge that uses the interface idx.
// Callers must hold b.mu.
func (b *BridgeManager) bridgeUsing(idx int) (string, bool) {
	for name, br := range b.bridges {
		if br.Interface1().Index() == idx || br.Interface2().Index() == idx {
			return name, true
		}
		if m := br.MonitorInterface(); m != nil && m.Index() == idx {
			return name, true
		}
	}
	return "", false
}

func (b *BridgeManager) Remove(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	ebpfBridge, ok := b.bridges[name]
	if !ok {
		return fmt.Errorf("bridge %s does not exist", name)
	}
	if err := ebpfBridge.Remove(); err != nil {
		return err
	}
	delete(b.bridges, name)
	return nil
}

// RemoveAll removes every bridge, e.g. when the service shuts down.
func (b *BridgeManager) RemoveAll() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var errs []error
	for name, br := range b.bridges {
		errs = append(errs, br.Remove())
		delete(b.bridges, name)
	}
	return errors.Join(errs...)
}

// List returns all bridges sorted by name.
func (b *BridgeManager) List() []BridgeDescription {
	b.mu.Lock()
	defer b.mu.Unlock()

	bridges := make([]BridgeDescription, 0, len(b.bridges))
	for name, br := range b.bridges {
		desc := BridgeDescription{
			Name:   name,
			Iface1: int32(br.Interface1().Index()),
			Iface2: int32(br.Interface2().Index()),
		}
		if m := br.MonitorInterface(); m != nil {
			idx := int32(m.Index())
			desc.Monitor = &idx
		}
		bridges = append(bridges, desc)
	}
	slices.SortFunc(bridges, func(a, b BridgeDescription) int {
		return strings.Compare(a.Name, b.Name)
	})
	return bridges
}
