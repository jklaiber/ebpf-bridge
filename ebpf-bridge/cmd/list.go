package cmd

import (
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/jklaiber/ebpf-bridge/pkg/printer"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all bridges",
	Run: func(cmd *cobra.Command, args []string) {
		messagingClient := messaging.NewMessagingClient()
		defer messagingClient.Close()

		msg := &messaging.ListCommand{}
		returnMsg, _ := messagingClient.ListBridges(msg)

		printer := &printer.PrettyPrinter{}
		printer.PrintBridgeDescriptions(returnMsg.Bridges)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
