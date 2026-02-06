package vault

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "tls",
	Short: "TLS Utils",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
