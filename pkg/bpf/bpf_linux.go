package bpf

import (
	"fmt"

	"github.com/cilium/ebpf"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS bpf ../../bpf/bridge.c

type BpfLinux struct{}

func (b *BpfLinux) ReadBpfObjects() (*bpfObjects, error) {
	obj := &bpfObjects{}
	ops := &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf",
		},
	}
	err := loadBpfObjects(obj, ops)
	if err != nil {
		return nil, fmt.Errorf("failed to load bpf objects: %w", err)
	}
	return obj, nil
}

func (b *BpfLinux) ReadBpfSpecs() (*ebpf.CollectionSpec, error) {
	specs, err := loadBpf()
	if err != nil {
		return nil, fmt.Errorf("failed to load bpf specs: %w", err)
	}
	return specs, nil
}
