package command

import (
	"github.com/jklaiber/ebpf-bridge/pkg/api"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/jklaiber/ebpf-bridge/pkg/printer"
)

type ListCommand struct {
	Command
	linkFactory hostlink.LinkFactory
}

func NewListCommand(linkFactory hostlink.LinkFactory, messagingClient messaging.Client) *ListCommand {
	return &ListCommand{
		Command:     Command{messagingClient: messagingClient},
		linkFactory: linkFactory,
	}
}

func (l *ListCommand) Execute() (string, error) {
	resp, err := l.messagingClient.ListBridges(&api.ListCommand{})
	if err != nil {
		return "", err
	}
	return printer.NewPrettyPrinter(l.linkFactory).PrintBridgeDescriptions(resp.GetBridges()), nil
}
