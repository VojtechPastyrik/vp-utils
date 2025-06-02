package root

import (
	"github.com/VojtechPastyrik/vp-utils/version"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "vp-utils",
	Short: "vp-utils, " + version.Version,
	Long:  "vp-utils, " + version.Version + "\n\nA collection of useful tools and utilities",
}
