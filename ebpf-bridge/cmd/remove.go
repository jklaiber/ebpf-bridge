package cmd

import (
	"fmt"

	"github.com/jklaiber/ebpf-bridge/pkg/command"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove bridge between two interfaces",
	Run: func(cmd *cobra.Command, args []string) {
		messagingClient := messaging.NewMessagingClient()
		defer messagingClient.Close()

		removeCommand := command.NewRemoveCommand(messagingClient, bridgeName)
		returnMsg, err := removeCommand.Execute()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(returnMsg)
	},
}

func init() {
	removeCmd.Flags().StringVarP(&bridgeName, "name", "n", "", "Name of the bridge to remove")
	removeCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(removeCmd)
}
