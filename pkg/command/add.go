package command

import (
	"errors"
	"fmt"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
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
		Command:         Command{messagingClient: messagingClient},
		hostLinkFactory: hostlinkFactory,
		BridgeName:      bridgeName,
		Iface1:          iface1,
		Iface2:          iface2,
		MonitorIface:    monitorIface,
	}
}

func (a *AddCommand) lookup(name string) (int32, error) {
	l, err := a.hostLinkFactory.NewLinkWithName(name)
	if err != nil {
		return 0, fmt.Errorf("interface %s: %w", name, err)
	}
	return int32(l.Index()), nil
}

func (a *AddCommand) Execute() (string, error) {
	msg := &api.AddCommand{Name: a.BridgeName}
	var err error
	if msg.Iface1, err = a.lookup(a.Iface1); err != nil {
		return "", err
	}
	if msg.Iface2, err = a.lookup(a.Iface2); err != nil {
		return "", err
	}
	if a.MonitorIface != "" {
		monitor, err := a.lookup(a.MonitorIface)
		if err != nil {
			return "", err
		}
		msg.Monitor = &monitor
	}

	resp, err := a.messagingClient.AddBridge(msg)
	if err != nil {
		return "", err
	}
	if !resp.GetSuccess() {
		return "", errors.New(resp.GetMessage())
	}
	return resp.GetMessage(), nil
}
