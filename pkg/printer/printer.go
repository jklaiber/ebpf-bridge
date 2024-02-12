//go:generate mockgen -source=printer.go -destination=mocks/printer_mock.go -package=mocks Printer
package printer

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
)

type Printer interface {
	PrintBridgeDescriptions(bridges []*messaging.BridgeDescription) string
}

type PrettyPrinter struct {
	linkFactory hostlink.LinkFactory
}

func NewPrettyPrinter(linkFactory hostlink.LinkFactory) *PrettyPrinter {
	return &PrettyPrinter{
		linkFactory: linkFactory,
	}
}

func (p *PrettyPrinter) PrintBridgeDescriptions(bridges []*messaging.BridgeDescription) string {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Iface1", "Iface2", "Monitor-Iface"})

	for _, bridge := range bridges {
		iface1, _ := p.linkFactory.NewLinkWithIndex(int(bridge.Iface1))
		iface2, _ := p.linkFactory.NewLinkWithIndex(int(bridge.Iface2))
		if bridge.Monitor == nil {
			t.AppendRow([]interface{}{bridge.Name, iface1.Name(), iface2.Name(), ""})
		} else {
			monitorIface, _ := p.linkFactory.NewLinkWithIndex(int(*bridge.Monitor))
			t.AppendRow([]interface{}{bridge.Name, iface1.Name(), iface2.Name(), monitorIface.Name()})
		}
		t.AppendSeparator()
	}
	return t.Render()
}
