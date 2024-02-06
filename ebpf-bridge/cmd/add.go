package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
)

var (
	iface1       string
	iface2       string
	monitorIface string
)

var linkCmd = &cobra.Command{
	Use:   "add",
	Short: "Add bridge between two interfaces",
	Run: func(cmd *cobra.Command, args []string) {
		// manager := bridge.NewBridgeManager()
		// manager.Add("test", iface1, iface2, monitorIface)
		messagingClient := messaging.NewMessagingClient()
		iface1Index, err := netlink.LinkByName(iface1)
		if err != nil {
			fmt.Println(err)
		}
		iface2Index, err := netlink.LinkByName(iface2)
		if err != nil {
			fmt.Println(err)
		}
		defer messagingClient.Close()
		msg := &messaging.AddCommand{
			Name:   "test",
			Iface1: int32(iface1Index.Attrs().Index),
			Iface2: int32(iface2Index.Attrs().Index),
		}
		returnMsg, err := messagingClient.AddBridge(msg)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(returnMsg)
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigChan:
			fmt.Println("Received signal, exiting")
		}
	},
}

func init() {
	linkCmd.Flags().StringVar(&iface1, "iface1", "", "First interface to connect")
	linkCmd.Flags().StringVar(&iface2, "iface2", "", "Second interface to connect")
	linkCmd.MarkFlagRequired("iface1")
	linkCmd.MarkFlagRequired("iface2")
	linkCmd.Flags().StringVar(&monitorIface, "monitor", "", "Monitoring interface")
	rootCmd.AddCommand(linkCmd)
}
