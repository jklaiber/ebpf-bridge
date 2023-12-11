package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log = logrus.New()

var rootCmd = &cobra.Command{
	Use:   "ebpf-bridge",
	Short: "ebpf-bridge is a tool to bridge network interfaces using eBPF",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	checkIsRoot()
}

func checkIsRoot() {
	if os.Getuid() != 0 {
		log.Fatal("You must be root to run this program.")
	}
}
