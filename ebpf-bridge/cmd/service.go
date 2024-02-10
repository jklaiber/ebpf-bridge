package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/service"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "start-service",
	Short: "Start the ebpf-bridge managing service",
	Run: func(cmd *cobra.Command, args []string) {
		checkIsRoot()
		linkFactory := hostlink.NewHostLinkFactory()
		service := service.NewEbpfBridgeService(linkFactory)
		service.Start()
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigChan:
			fmt.Println("Received signal, exiting")
			service.Stop()
		}
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}

func checkIsRoot() {
	if os.Getuid() != 0 {
		log.Fatal("The ebpf-bridge service must be run as root.")
	}
}
