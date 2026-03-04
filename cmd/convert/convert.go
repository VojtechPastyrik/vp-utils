package convert

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "convert",
	Short: "Data format conversion utilities",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
