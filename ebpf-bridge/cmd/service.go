package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jklaiber/ebpf-bridge/pkg/bpf"
	"github.com/jklaiber/ebpf-bridge/pkg/bridge"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/manager"
	"github.com/jklaiber/ebpf-bridge/pkg/service"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "start-service",
	Short: "Start the ebpf-bridge managing service",
	Run: func(cmd *cobra.Command, args []string) {
		checkIsRoot()
		bpf := bpf.NewBpfLinux()
		linkFactory := hostlink.NewHostLinkFactory()
		bridgeFactory := bridge.NewEbpfBridgeFactory(bpf)
		bridgeManager := manager.NewBridgeManager(linkFactory, bridgeFactory)
		service := service.NewEbpfBridgeService(bridgeManager)
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
