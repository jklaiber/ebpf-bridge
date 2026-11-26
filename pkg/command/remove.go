package command

import (
	"errors"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
)

type RemoveCommand struct {
	Command
	Name string
}

func NewRemoveCommand(messagingClient messaging.Client, name string) *RemoveCommand {
	return &RemoveCommand{
		Command: Command{messagingClient: messagingClient},
		Name:    name,
	}
}

func (r *RemoveCommand) Execute() (string, error) {
	resp, err := r.messagingClient.RemoveBridge(&api.RemoveCommand{Name: r.Name})
	if err != nil {
		return "", err
	}
	if !resp.GetSuccess() {
		return "", errors.New(resp.GetMessage())
	}
	return resp.GetMessage(), nil
}
