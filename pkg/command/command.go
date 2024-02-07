package command

import "github.com/jklaiber/ebpf-bridge/pkg/messaging"

type CommandInterface interface {
	Execute() (string, error)
}

type Command struct {
	messagingClient messaging.Client
}
