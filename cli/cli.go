package cli

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.AddCommand(initServerCmd())
	return cmd
}
