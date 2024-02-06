package cmd

import (
	"fmt"

	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove bridge between two interfaces",
	Run: func(cmd *cobra.Command, args []string) {
		messagingClient := messaging.NewMessagingClient()
		defer messagingClient.Close()
		msg := &messaging.RemoveCommand{
			Name: "test",
		}
		returnMsg, err := messagingClient.RemoveBridge(msg)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(returnMsg)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
