package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jklaiber/ebpf-bridge/pkg/command"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all bridges",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withClient(func(client messaging.Client) error {
				out, err := command.NewListCommand(hostlink.NewHostLinkFactory(), client).Execute()
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), out)
				return err
			})
		},
	}
}

func init() {
	rootCmd.AddCommand(newListCmd())
}
