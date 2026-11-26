package cmd

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jklaiber/ebpf-bridge/pkg/bpf"
	"github.com/jklaiber/ebpf-bridge/pkg/bridge"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/linker"
	"github.com/jklaiber/ebpf-bridge/pkg/logging"
	"github.com/jklaiber/ebpf-bridge/pkg/manager"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
	"github.com/jklaiber/ebpf-bridge/pkg/service"
)

func newServiceCmd() *cobra.Command {
	var logLevel string
	c := &cobra.Command{
		Use:   "start-service",
		Short: "Start the ebpf-bridge service",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if os.Geteuid() != 0 {
				return errors.New("the ebpf-bridge service must be run as root")
			}
			level, err := logrus.ParseLevel(logLevel)
			if err != nil {
				return err
			}
			logging.DefaultLogger.SetLevel(level)

			bridgeManager := manager.NewBridgeManager(
				hostlink.NewHostLinkFactory(),
				bridge.NewEbpfBridgeFactory(bpf.NewKernelLoader(), linker.NewXdpLinkerFactory()),
			)
			svc := service.NewEbpfBridgeService(bridgeManager, messaging.NewMessagingServer(bridgeManager, socketPath))
			if err := svc.Start(); err != nil {
				return err
			}

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			<-sig
			return svc.Stop()
		},
	}
	c.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	return c
}

func init() {
	rootCmd.AddCommand(newServiceCmd())
}
