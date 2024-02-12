//go:generate mockgen -source=bpf.go -destination=bpf_mock.go -package=bpf Bpf
package bpf

import (
	"github.com/cilium/ebpf"
)

type Bpf interface {
	ReadBpfObjects() (*bpfObjects, error)
	ReadBpfSpecs() (*ebpf.CollectionSpec, error)
	LoadPinnedMap(mapPath string) (*ebpf.Map, error)
}
