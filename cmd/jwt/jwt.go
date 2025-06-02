package jwt

import (
	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "jwt",
	Short: "JWT Utils",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
