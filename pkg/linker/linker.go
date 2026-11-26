// Package linker attaches XDP programs to network interfaces.
package linker

//go:generate go tool mockgen -source=linker.go -destination=mocks/linker_mock.go -package=mocks

import (
	"errors"
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

// Factory creates a Linker that attaches program to iface.
type Factory func(iface hostlink.Link, program *ebpf.Program) Linker

// NewXdpLinkerFactory returns a Factory that creates XdpLinkers.
func NewXdpLinkerFactory() Factory {
	return func(iface hostlink.Link, program *ebpf.Program) Linker {
		return NewXdpLinker(iface, program)
	}
}

// XdpLinker attaches a program in generic (SKB) mode, which works on every
// driver including veth.
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
	if x.program == nil {
		return errors.New("cannot attach a nil program")
	}
	log.Infof("Attaching XDP program to %s", x.iface.Name())
	l, err := link.AttachXDP(link.XDPOptions{
		Program:   x.program,
		Interface: x.iface.Index(),
		Flags:     link.XDPGenericMode,
	})
	if err != nil {
		return fmt.Errorf("failed to attach XDP program to interface %s: %w", x.iface.Name(), err)
	}
	x.link = l
	return nil
}

func (x *XdpLinker) Detach() error {
	if x.link == nil {
		return nil
	}
	log.Infof("Detaching XDP program from %s", x.iface.Name())
	if err := x.link.Close(); err != nil {
		return fmt.Errorf("failed to detach XDP program from interface %s: %w", x.iface.Name(), err)
	}
	x.link = nil
	return nil
}
