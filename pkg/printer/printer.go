//go:generate mockgen -source=printer.go -destination=mocks/printer_mock.go -package=mocks Printer
package printer

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jklaiber/ebpf-bridge/pkg/api"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
)

type Printer interface {
	PrintBridgeDescriptions(bridges []*api.BridgeDescription) string
}

type PrettyPrinter struct {
	linkFactory hostlink.LinkFactory
}

func NewPrettyPrinter(linkFactory hostlink.LinkFactory) *PrettyPrinter {
	return &PrettyPrinter{
		linkFactory: linkFactory,
	}
}

func (p *PrettyPrinter) PrintBridgeDescriptions(bridges []*api.BridgeDescription) string {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Iface1", "Iface2", "Monitor-Iface"})

	if len(bridges) == 0 {
		fmt.Println("No bridges found")
		return "No bridges found"
	}

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
