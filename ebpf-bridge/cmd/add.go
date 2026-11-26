package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jklaiber/ebpf-bridge/pkg/command"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
)

func newAddCmd() *cobra.Command {
	var name, iface1, iface2, monitor string
	c := &cobra.Command{
		Use:     "add",
		Short:   "Add a bridge between two interfaces",
		Example: "  ebpf-bridge add --name br0 --iface1 eth0 --iface2 eth1 --monitor eth2",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withClient(func(client messaging.Client) error {
				msg, err := command.NewAddCommand(hostlink.NewHostLinkFactory(), client, name, iface1, iface2, monitor).Execute()
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), msg)
				return err
			})
		},
	}
	c.Flags().StringVarP(&name, "name", "n", "", "Name of the bridge")
	c.Flags().StringVar(&iface1, "iface1", "", "First interface to connect")
	c.Flags().StringVar(&iface2, "iface2", "", "Second interface to connect")
	c.Flags().StringVar(&monitor, "monitor", "", "Optional interface that receives a copy of all bridged traffic")
	for _, f := range []string{"name", "iface1", "iface2"} {
		_ = c.MarkFlagRequired(f)
	}
	return c
}

func init() {
	rootCmd.AddCommand(newAddCmd())
}
