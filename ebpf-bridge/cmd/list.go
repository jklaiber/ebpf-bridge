package cmd

import (
	"fmt"

	"github.com/jklaiber/ebpf-bridge/pkg/command"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all bridges",
	Run: func(cmd *cobra.Command, args []string) {
		messagingClient := messaging.NewMessagingClient()
		defer messagingClient.Close()

		listCommand := command.NewListCommand(messagingClient)
		_, err := listCommand.Execute()

		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
