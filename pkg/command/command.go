// Package command implements the client side of the CLI commands.
package command

import "github.com/jklaiber/ebpf-bridge/pkg/messaging"

type Command struct {
	messagingClient messaging.Client
}
