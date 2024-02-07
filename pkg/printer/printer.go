package printer

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/vishvananda/netlink"
)

type Printer interface {
	PrintBridgeDescriptions(bridges []*messaging.BridgeDescription)
}

type PrettyPrinter struct{}

func (p *PrettyPrinter) PrintBridgeDescriptions(bridges []*messaging.BridgeDescription) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Iface1", "Iface2", "Monitor-Iface"})

	for _, bridge := range bridges {
		if bridge.Monitor == nil {
			t.AppendRow([]interface{}{bridge.Name, GetIfaceNameForIndex(bridge.Iface1), GetIfaceNameForIndex(bridge.Iface2), ""})
		} else {
			t.AppendRow([]interface{}{bridge.Name, GetIfaceNameForIndex(bridge.Iface1), GetIfaceNameForIndex(bridge.Iface2), GetIfaceNameForIndex(*bridge.Monitor)})
		}
		t.AppendSeparator()
	}
	t.Render()
}

func GetIfaceNameForIndex(index int32) string {
	iface, err := netlink.LinkByIndex(int(index))
	if err != nil {
		return ""
	}
	return iface.Attrs().Name
}
