package command

import (
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/jklaiber/ebpf-bridge/pkg/printer"
)

type ListCommand struct {
	Command
}

func NewListCommand(messagingClient messaging.Client) *ListCommand {
	return &ListCommand{
		Command: Command{
			messagingClient: messagingClient,
		},
	}
}

func (l *ListCommand) Execute() (string, error) {
	returnMsg, err := l.messagingClient.ListBridges(&messaging.ListCommand{})
	if err != nil {
		return "", err
	}
	printer := &printer.PrettyPrinter{}
	return printer.PrintBridgeDescriptions(returnMsg.Bridges), nil
}
