package gpg

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "gpg",
	Short: "GPG key management utilities",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
