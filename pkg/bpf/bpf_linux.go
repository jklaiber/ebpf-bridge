package bpf

import "fmt"

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS bpf ../../bpf/bridge.c

type BpfLinux struct{}

func (b *BpfLinux) ReadBpfObjects() (*bpfObjects, error) {
	obj := &linkObjects{}
	ops := &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf",
		},
	}
	err := loadBpfObjects(obj, ops)
	if err != nil {
		return nil, fmt.Errorf("failed to load link objects: %w", err)
	}
	return obj, nil
}

func (b *BpfLinux) ReadBpfSpecs() (*ebpf.CollectionSpecs, error) {
	specs, err := loadBpf()
	if err != nil {
		return nil, fmt.Errorf("failed to load link specs: %w", err)
	}
	return specs, nil
}
