// Package bpf loads the XDP bridge program and gives access to its port map.
package bpf

//go:generate go tool bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS bpf ../../bpf/bridge.c
//go:generate go tool mockgen -source=bpf.go -destination=mocks/bpf_mock.go -package=mocks

import (
	"fmt"

	"github.com/cilium/ebpf"
)

// Port slots in the devmap of a bridge.
const (
	SlotIface1  uint32 = 0
	SlotIface2  uint32 = 1
	SlotMonitor uint32 = 2
)

// Objects is one loaded instance of the bridge program and its port map.
type Objects interface {
	// Program returns the XDP program to attach to the bridged interfaces.
	Program() *ebpf.Program
	// SetPort stores the interface index ifindex in the given port slot.
	SetPort(slot uint32, ifindex int) error
	// Close releases the program and the map.
	Close() error
}

// Loader loads a fresh instance of the bridge program.
type Loader interface {
	Load() (Objects, error)
}

// KernelLoader loads the embedded program into the running kernel.
type KernelLoader struct{}

func NewKernelLoader() *KernelLoader {
	return &KernelLoader{}
}

func (KernelLoader) Load() (Objects, error) {
	objs := &bpfObjects{}
	if err := loadBpfObjects(objs, nil); err != nil {
		return nil, fmt.Errorf("failed to load bpf objects: %w", err)
	}
	return &kernelObjects{objs: objs}, nil
}

type kernelObjects struct {
	objs *bpfObjects
}

func (k *kernelObjects) Program() *ebpf.Program {
	return k.objs.XdpBridge
}

func (k *kernelObjects) SetPort(slot uint32, ifindex int) error {
	if err := k.objs.Devmap.Put(slot, uint32(ifindex)); err != nil {
		return fmt.Errorf("failed to set port %d to ifindex %d: %w", slot, ifindex, err)
	}
	return nil
}

func (k *kernelObjects) Close() error {
	return k.objs.Close()
}
