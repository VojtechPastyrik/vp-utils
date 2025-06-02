package vault

import (
	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "vault",
	Short: "vault Utils",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
