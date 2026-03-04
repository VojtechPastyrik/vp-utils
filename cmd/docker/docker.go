package docker

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker cleanup and monitoring utilities",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
