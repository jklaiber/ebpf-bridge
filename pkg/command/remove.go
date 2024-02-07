package command

import "github.com/jklaiber/ebpf-bridge/pkg/messaging"

type RemoveCommand struct {
	Command
	Name string
}

func NewRemoveCommand(messagingClient messaging.Client, name string) *RemoveCommand {
	return &RemoveCommand{
		Command: Command{
			messagingClient: messagingClient,
		},
		Name: name,
	}
}

func (r *RemoveCommand) Execute() (string, error) {
	msg := &messaging.RemoveCommand{
		Name: r.Name,
	}
	returnMsg, err := r.messagingClient.RemoveBridge(msg)
	if err != nil {
		return "", err
	}
	return returnMsg.Message, nil
}
