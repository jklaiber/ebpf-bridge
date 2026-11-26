// Package cmd implements the ebpf-bridge command line interface.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
)

// version is set at build time via -ldflags "-X .../cmd.version=...".
var version = "dev"

var socketPath string

var rootCmd = &cobra.Command{
	Use:           "ebpf-bridge",
	Short:         "ebpf-bridge bridges network interfaces using eBPF",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&socketPath, "socket", messaging.DefaultSocket, "Unix socket of the ebpf-bridge service")
}

// Execute runs the CLI and exits with a non-zero code on failure.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// withClient connects to the service and runs fn with the client.
func withClient(fn func(messaging.Client) error) error {
	client, err := messaging.NewMessagingClient(socketPath)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()
	return fn(client)
}
