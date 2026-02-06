package azure

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "azure",
	Aliases: []string{"az"},
	Short:   "Custom Azure CLI Utils, which are not available in the official Azure CLI",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
