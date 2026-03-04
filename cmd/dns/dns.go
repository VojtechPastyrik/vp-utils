package dns

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS lookup utilities",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
