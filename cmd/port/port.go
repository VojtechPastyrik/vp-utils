package port

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "port",
	Short: "Port check and scan utilities",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
