// Package printer renders bridges for the CLI.
package printer

//go:generate go tool mockgen -source=printer.go -destination=mocks/printer_mock.go -package=mocks

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
)

// NoBridges is printed when the service manages no bridges.
const NoBridges = "No bridges found"

type Printer interface {
	PrintBridgeDescriptions(bridges []*api.BridgeDescription) string
}

type PrettyPrinter struct {
	linkFactory hostlink.LinkFactory
}

func NewPrettyPrinter(linkFactory hostlink.LinkFactory) *PrettyPrinter {
	return &PrettyPrinter{linkFactory: linkFactory}
}

func (p *PrettyPrinter) PrintBridgeDescriptions(bridges []*api.BridgeDescription) string {
	if len(bridges) == 0 {
		return NoBridges
	}
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Name", "Iface1", "Iface2", "Monitor-Iface"})
	for _, b := range bridges {
		monitor := ""
		if b.Monitor != nil {
			monitor = p.ifaceName(*b.Monitor)
		}
		t.AppendRow(table.Row{b.GetName(), p.ifaceName(b.GetIface1()), p.ifaceName(b.GetIface2()), monitor})
		t.AppendSeparator()
	}
	return t.Render()
}

// ifaceName resolves an interface index to its name. The CLI may run in a
// different network namespace than the service or the interface may be gone,
// so fall back to the index.
func (p *PrettyPrinter) ifaceName(index int32) string {
	l, err := p.linkFactory.NewLinkWithIndex(int(index))
	if err != nil || l == nil {
		return fmt.Sprintf("#%d", index)
	}
	return l.Name()
}
