package cpu

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cpu",
	Short: "CPU Utils",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
