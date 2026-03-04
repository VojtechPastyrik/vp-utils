package cert

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cert",
	Short: "TLS certificate inspection utilities",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
