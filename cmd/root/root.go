package root

import (
	"github.com/VojtechPastyrik/vpd/version"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "vpd",
	Short: "vpd, " + version.Version,
	Long:  "vpd, " + version.Version + "\n\nA collection of useful tools and utilities",
}
