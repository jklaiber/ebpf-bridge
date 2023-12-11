package link

import "fmt"

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS link ../../../bpf/bpf_link.c

type Link interface {
	ReadLinkBpfObjects() (*linkObjects, error)
	ReadLinkBpfSpecs() (*ebpf.CollectionSpecs, error)
}

type RealLinkReader struct{}

func (r *RealLinkReader) ReadLinkBpfObjects() (*linkObjects, error) {
	obj := &linkObjects{}
	ops := &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf",
		},
	}
	err := loadLinkObjects(obj, ops)
	if err != nil {
		return nil, fmt.Errorf("failed to load link objects: %w", err)
	}
	return obj, nil
}

func (r *RealLinkReader) ReadLinkBpfSpecs() (*ebpf.CollectionSpecs, error) {
	specs, err := loadLink()
	if err != nil {
		return nil, fmt.Errorf("failed to load link specs: %w", err)
	}
	return specs, nil
}
