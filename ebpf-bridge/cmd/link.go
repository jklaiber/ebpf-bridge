package cmd

import "github.com/spf13/cobra"

var (
	iface1       string
	iface2       string
	monitorIface string
)

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Connect two network interfaces using eBPF",
}

func init() {
	linkCmd.Flags().StringVar(&iface1, "iface1", "", "First interface to connect")
	linkCmd.Flags().StringVar(&iface2, "iface2", "", "Second interface to connect")
	linkCmd.MarkFlagRequired("iface1")
	linkCmd.MarkFlagRequired("iface2")
	linkCmd.Flags().StringVar(&monitorIface, "monitor", "", "Monitoring interface")
	rootCmd.AddCommand(linkCmd)
}
