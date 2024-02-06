package cmd

import (
	"github.com/jklaiber/ebpf-bridge/pkg/logging"
	"github.com/spf13/cobra"
)

var log = logging.DefaultLogger.WithField("subsystem", "cli")

var rootCmd = &cobra.Command{
	Use:   "ebpf-bridge",
	Short: "ebpf-bridge is a tool to bridge network interfaces using eBPF",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
