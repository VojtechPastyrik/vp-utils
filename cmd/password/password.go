package password

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "password",
	Aliases: []string{"pwd", "secret"},
	Short:   "Password utilities",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
