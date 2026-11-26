package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jklaiber/ebpf-bridge/pkg/command"
	"github.com/jklaiber/ebpf-bridge/pkg/messaging"
)

func newRemoveCmd() *cobra.Command {
	var name string
	c := &cobra.Command{
		Use:   "remove",
		Short: "Remove a bridge",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withClient(func(client messaging.Client) error {
				msg, err := command.NewRemoveCommand(client, name).Execute()
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), msg)
				return err
			})
		},
	}
	c.Flags().StringVarP(&name, "name", "n", "", "Name of the bridge to remove")
	_ = c.MarkFlagRequired("name")
	return c
}

func init() {
	rootCmd.AddCommand(newRemoveCmd())
}
