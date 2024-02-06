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
			Name: bridgeName,
		}
		returnMsg, err := messagingClient.RemoveBridge(msg)
		if err != nil {
			fmt.Println(err)
		}

		if returnMsg.Success {
			fmt.Printf("Bridge %s removed\n", bridgeName)
		} else {
			fmt.Printf("Failed to remove bridge %s\n", bridgeName)
		}
	},
}

func init() {
	removeCmd.Flags().StringVarP(&bridgeName, "name", "n", "", "Name of the bridge to remove")
	removeCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(removeCmd)
}
