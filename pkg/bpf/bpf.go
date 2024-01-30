package bpf

type Bpf interface {
	ReadBpfObjects() (*bpfObjects, error)
	ReadBpfSpecs() (*ebpf.CollectionSpecs, error)
}
