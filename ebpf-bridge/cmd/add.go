package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jklaiber/ebpf-bridge/pkg/manager"
	"github.com/spf13/cobra"
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
		manager := manager.NewBridgeManager()
		manager.Add("test", iface1, iface2, monitorIface)
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigChan:
			fmt.Println("Received signal, exiting")
			manager.Remove("test")
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
