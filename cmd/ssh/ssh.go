package ssh

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH utilities",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
