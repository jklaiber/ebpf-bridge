package bpf

import (
	"github.com/cilium/ebpf"
)

type Bpf interface {
	ReadBpfObjects() (*bpfObjects, error)
	ReadBpfSpecs() (*ebpf.CollectionSpec, error)
}
