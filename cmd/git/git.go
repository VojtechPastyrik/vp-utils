package jwt

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "git",
	Short: "GIT Utils",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
