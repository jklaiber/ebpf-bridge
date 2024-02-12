package command

import (
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
)

type AddCommand struct {
	Command
	hostLinkFactory hostlink.LinkFactory
	BridgeName      string
	Iface1          string
	Iface2          string
	MonitorIface    string
}

func NewAddCommand(hostlinkFactory hostlink.LinkFactory, messagingClient messaging.Client, bridgeName, iface1, iface2, monitorIface string) *AddCommand {
	return &AddCommand{
		Command: Command{
			messagingClient: messagingClient,
		},
		hostLinkFactory: hostlinkFactory,
		BridgeName:      bridgeName,
		Iface1:          iface1,
		Iface2:          iface2,
		MonitorIface:    monitorIface,
	}
}

func (a *AddCommand) Execute() (string, error) {
	iface1Index, err := a.hostLinkFactory.NewLinkWithName(a.Iface1)
	if err != nil {
		return "", err
	}

	iface2Index, err := a.hostLinkFactory.NewLinkWithName(a.Iface2)
	if err != nil {
		return "", err
	}

	msg := &messaging.AddCommand{
		Name: a.BridgeName,
	}

	if a.MonitorIface != "" {
		monitorIfaceIndex, err := a.hostLinkFactory.NewLinkWithName(a.MonitorIface)
		if err != nil {
			return "", err
		}
		msg.Iface1 = int32(iface1Index.Index())
		msg.Iface2 = int32(iface2Index.Index())
		monitorValue := int32(monitorIfaceIndex.Index())
		msg.Monitor = &monitorValue
	} else {
		msg.Iface1 = int32(iface1Index.Index())
		msg.Iface2 = int32(iface2Index.Index())
	}

	returnMsg, err := a.messagingClient.AddBridge(msg)
	if err != nil {
		return "", err
	}

	return returnMsg.Message, nil
}
