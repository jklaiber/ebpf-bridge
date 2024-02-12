package cmd

import (
	"fmt"

	"github.com/jklaiber/ebpf-bridge/pkg/command"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/spf13/cobra"
)

var (
	bridgeName   string
	iface1       string
	iface2       string
	monitorIface string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add bridge between two interfaces",
	Run: func(cmd *cobra.Command, args []string) {
		hostlinkFactory := hostlink.NewHostLinkFactory()
		messagingClient := messaging.NewMessagingClient()
		defer messagingClient.Close()
		addCommand := command.NewAddCommand(hostlinkFactory, messagingClient, bridgeName, iface1, iface2, monitorIface)
		returnMsg, err := addCommand.Execute()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(returnMsg)
	},
}

func init() {
	addCmd.Flags().StringVar(&bridgeName, "name", "", "Name of the bridge")
	addCmd.MarkFlagRequired("name")
	addCmd.Flags().StringVar(&iface1, "iface1", "", "First interface to connect")
	addCmd.MarkFlagRequired("iface1")
	addCmd.Flags().StringVar(&iface2, "iface2", "", "Second interface to connect")
	addCmd.MarkFlagRequired("iface2")
	addCmd.Flags().StringVar(&monitorIface, "monitor", "", "Monitoring interface")
	rootCmd.AddCommand(addCmd)
}
