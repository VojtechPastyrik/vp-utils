package youtrack

import (
	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "youtrack",
	Aliases: []string{"yt"},
	Short:   "YouTrack CLI Utils",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}
