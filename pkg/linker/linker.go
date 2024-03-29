//go:generate mockgen -source=linker.go -destination=mocks/linker_mock.go -package=mocks Linker
package linker

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/logging"
)

var log = logging.DefaultLogger.WithField("subsystem", "linker")

type Linker interface {
	Attach() error
	Detach() error
}

type XdpLinker struct {
	iface   hostlink.Link
	program *ebpf.Program
	link    link.Link
}

func NewXdpLinker(iface hostlink.Link, program *ebpf.Program) *XdpLinker {
	return &XdpLinker{
		iface:   iface,
		program: program,
	}
}

func (x *XdpLinker) Attach() error {
	log.Infof("Attaching XDP program to %s", x.iface.Name())
	if x.program == nil {
		return fmt.Errorf("cannot attach a nil program")
	}
	link, err := link.AttachXDP(link.XDPOptions{
		Program:   x.program,
		Interface: x.iface.Index(),
		Flags:     link.XDPGenericMode,
	})
	if err != nil {
		return fmt.Errorf("failed to attach XDP program to interface %s: %w", x.iface.Name(), err)
	}
	x.link = link
	return nil
}

func (x *XdpLinker) Detach() error {
	log.Infof("Detaching XDP program from %s", x.iface.Name())
	err := x.link.Close()
	if err != nil {
		return fmt.Errorf("failed to detach XDP program: %w", err)
	}
	return nil
}
