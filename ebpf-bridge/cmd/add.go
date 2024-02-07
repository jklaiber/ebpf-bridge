package cmd

import (
	"fmt"

	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
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
		messagingClient := messaging.NewMessagingClient()
		defer messagingClient.Close()

		iface1Index, err := netlink.LinkByName(iface1)
		if err != nil {
			fmt.Println(err)
		}
		iface2Index, err := netlink.LinkByName(iface2)
		if err != nil {
			fmt.Println(err)
		}

		msg := &messaging.AddCommand{
			Name: bridgeName,
		}

		if monitorIface != "" {
			monitorIfaceIndex, err := netlink.LinkByName(monitorIface)
			if err != nil {
				fmt.Println(err)
			}
			msg.Iface1 = int32(iface1Index.Attrs().Index)
			msg.Iface2 = int32(iface2Index.Attrs().Index)
			monitorValue := int32(monitorIfaceIndex.Attrs().Index)
			msg.Monitor = &monitorValue
		} else {
			msg.Iface1 = int32(iface1Index.Attrs().Index)
			msg.Iface2 = int32(iface2Index.Attrs().Index)
		}
		returnMsg, _ := messagingClient.AddBridge(msg)
		fmt.Println(returnMsg.Message)
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
