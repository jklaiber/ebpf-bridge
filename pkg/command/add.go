package command

import (
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/vishvananda/netlink"
)

type AddCommand struct {
	Command
	BridgeName   string
	Iface1       string
	Iface2       string
	MonitorIface string
}

func NewAddCommand(messagingClient messaging.Client, bridgeName, iface1, iface2, monitorIface string) *AddCommand {
	return &AddCommand{
		Command: Command{
			messagingClient: messagingClient,
		},
		BridgeName:   bridgeName,
		Iface1:       iface1,
		Iface2:       iface2,
		MonitorIface: monitorIface,
	}
}

func (a *AddCommand) Execute() (string, error) {
	iface1Index, err := netlink.LinkByName(a.Iface1)
	if err != nil {
		return "", err
	}

	iface2Index, err := netlink.LinkByName(a.Iface2)
	if err != nil {
		return "", err
	}

	msg := &messaging.AddCommand{
		Name: a.BridgeName,
	}

	if a.MonitorIface != "" {
		monitorIfaceIndex, err := netlink.LinkByName(a.MonitorIface)
		if err != nil {
			return "", err
		}
		msg.Iface1 = int32(iface1Index.Attrs().Index)
		msg.Iface2 = int32(iface2Index.Attrs().Index)
		monitorValue := int32(monitorIfaceIndex.Attrs().Index)
		msg.Monitor = &monitorValue
	} else {
		msg.Iface1 = int32(iface1Index.Attrs().Index)
		msg.Iface2 = int32(iface2Index.Attrs().Index)
	}

	returnMsg, err := a.messagingClient.AddBridge(msg)
	if err != nil {
		return "", err
	}

	return returnMsg.Message, nil
}
